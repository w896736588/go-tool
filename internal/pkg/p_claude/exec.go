package p_claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cast"
)

// maxScanTokenSize bufio.Scanner 最大缓冲区大小（10MB）。
// stream-json 输出中单行可能远超默认 64KB（尤其是工具调用结果）。
const maxScanTokenSize = 10 * 1024 * 1024

// ptyResult PTY 启动结果，由各平台 startClaude 实现返回。
type ptyResult struct {
	reader  io.ReadCloser
	pid     int
	waitFn  func() (int, error)
	closeFn func()
}

// RunClaudeStream 执行 claude 命令并逐行推送解析后的消息。
// callback 每收到一行 stream-json 时同步调用。
// 返回 sessionID（首行 init 中提取）和 error。
func RunClaudeStream(ctx context.Context, cfg RunConfig, callback func(msg StreamMessage)) (string, error) {
	args := buildArgs(cfg)
	env := buildEnv(cfg)

	log.Printf("[claude-exec] 启动进程, dir=%s model=%s", cfg.WorkingDir, cfg.Model)

	result, err := startClaude(ctx, args, cfg.WorkingDir, env)
	if err != nil {
		log.Printf("[claude-exec] 启动失败: %v", err)
		return ``, fmt.Errorf("claude start failed: %w", err)
	}
	defer result.closeFn()
	log.Printf("[claude-exec] 进程已启动, pid=%d", result.pid)

	sessionID := ``
	sessionExtracted := false
	lineCount := 0
	sc := bufio.NewScanner(result.reader)
	sc.Buffer(make([]byte, maxScanTokenSize), maxScanTokenSize)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == `` {
			continue
		}
		lineCount++
		if lineCount <= 3 {
			log.Printf("[claude-exec] 收到第%d行(len=%d): %.200s", lineCount, len(line), line)
		}
		msg := parseLine(line)
		callback(msg)

		if !sessionExtracted && cfg.SessionID == `` {
			if sid := extractSessionIDFromLine(line); sid != `` {
				sessionID = sid
				log.Printf("[claude-exec] 提取到 session_id=%s", sid)
				sessionExtracted = true
			}
		}
	}

	log.Printf("[claude-exec] scanner 循环结束, 总行数=%d", lineCount)
	if err := sc.Err(); err != nil {
		log.Printf("[claude-exec] scanner 错误: %v", err)
		return sessionID, fmt.Errorf("scan stdout: %w", err)
	}

	exitCode, waitErr := result.waitFn()
	log.Printf("[claude-exec] 进程结束, exitCode=%d waitErr=%v", exitCode, waitErr)
	if waitErr != nil {
		return sessionID, fmt.Errorf("claude exited with error: %w", waitErr)
	}
	if exitCode != 0 {
		return sessionID, fmt.Errorf("claude exited with code %d", exitCode)
	}
	return sessionID, nil
}

// buildArgs 构建 claude 命令行参数。
func buildArgs(cfg RunConfig) []string {
	args := []string{}
	if cfg.SessionID != `` {
		args = append(args, `--resume`, cfg.SessionID)
	}
	args = append(args,
		`-p`, cfg.Prompt,
		`--add-dir`, cfg.WorkingDir,
		`--output-format`, `stream-json`,
		`--include-partial-messages`,
		`--verbose`,
		`--permission-mode`, `bypassPermissions`,
	)
	if cfg.Model != `` {
		args = append(args, `--model`, cfg.Model)
	}
	if cfg.UserDataDir != `` {
		args = append(args, `--user-data-dir`, cfg.UserDataDir)
	}
	return args
}

// buildEnv 构建环境变量（在系统环境基础上追加）。
func buildEnv(cfg RunConfig) []string {
	env := os.Environ()
	if cfg.BaseURL != `` {
		env = append(env, `ANTHROPIC_BASE_URL=`+cfg.BaseURL)
	}
	if cfg.APIKey != `` {
		env = append(env, `ANTHROPIC_API_KEY=`+cfg.APIKey)
	}
	return env
}

// buildWindowsCmdLine 将参数列表转为 Windows CreateProcess 命令行字符串。
func buildWindowsCmdLine(args []string) string {
	var b strings.Builder
	for i, arg := range args {
		if i > 0 {
			b.WriteByte(' ')
		}
		if strings.ContainsAny(arg, " \t\n\v\"") {
			b.WriteByte('"')
			b.WriteString(strings.ReplaceAll(arg, `"`, `\"`))
			b.WriteByte('"')
		} else {
			b.WriteString(arg)
		}
	}
	return b.String()
}

// parseLine 解析单行 stream-json 为 StreamMessage。
func parseLine(line string) StreamMessage {
	msg := StreamMessage{RawJSON: line}
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		msg.Type = `raw_text`
		msg.Data = map[string]any{`text`: line}
		if errJSON, e := json.Marshal(msg); e == nil {
			msg.RawJSON = string(errJSON)
		}
		return msg
	}
	msg.Type = cast.ToString(raw[`type`])
	msg.Subtype = cast.ToString(raw[`subtype`])
	if event, ok := raw[`event`].(map[string]any); ok {
		msg.Event = cast.ToString(event[`type`])
	}
	msg.Data = raw
	return msg
}

// extractSessionIDFromLine 从 stream-json 行提取 session_id。
func extractSessionIDFromLine(line string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return ``
	}
	if cast.ToString(data[`type`]) != `system` || cast.ToString(data[`subtype`]) != `init` {
		return ``
	}
	return cast.ToString(data[`session_id`])
}
