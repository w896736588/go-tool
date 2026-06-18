package worker

import (
	"context"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_claude"
	"dev_tool/internal/pkg/p_codex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

// AgentCliResult Agent CLI 执行结果。
type AgentCliResult struct {
	Content   string // 最终汇总文本
	Success   bool   // 是否成功完成
	SessionID string // Agent CLI 会话 ID（可用于续接）
}

// agentCliTimeout Agent CLI 最大执行时间。
const agentCliTimeout = 10 * time.Minute

// getWorkingDir 获取当前工作目录，作为 Agent CLI 的工作目录。
func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return `.`
	}
	return dir
}

// RunAgentCli 使用 Agent CLI 执行复杂任务。
// 从 tbl_agent_cli 加载配置，根据类型（claude-code-cli / codex-cli）调用对应的 CLI。
func RunAgentCli(db *common.CSqlite, agentCliId int, prompt string) *AgentCliResult {
	// 加载 AgentCli 配置
	cliItem, err := loadAgentCliConfig(db, agentCliId)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler-agent-cli] 加载 AgentCli 配置失败 %s`, err.Error())
		return &AgentCliResult{Content: fmt.Sprintf(`Agent CLI 配置加载失败：%s`, err.Error()), Success: false}
	}
	if cliItem == nil {
		return &AgentCliResult{Content: `Agent CLI 未找到对应配置`, Success: false}
	}

	gstool.FmtPrintlnLogTime(`[butler-agent-cli] 启动执行 type=%s name=%s`, cliItem.Type, cliItem.Name)

	switch cliItem.Type {
	case define.AgentCliTypeClaudeCodeCli:
		return runClaudeCli(cliItem, prompt)
	case define.AgentCliTypeCodexCli:
		return runCodexCli(cliItem, prompt)
	default:
		return &AgentCliResult{Content: fmt.Sprintf(`不支持的 Agent CLI 类型：%s`, cliItem.Type), Success: false}
	}
}

// loadAgentCliConfig 从数据库加载 AgentCli 配置项。
func loadAgentCliConfig(db *common.CSqlite, agentCliId int) (*define.AgentCliItem, error) {
	row, err := db.Client.QuickQuery(`tbl_agent_cli`, `*`, map[string]any{
		`id`: agentCliId,
	}).One()
	if err != nil {
		return nil, fmt.Errorf(`查询 tbl_agent_cli 失败: %w`, err)
	}
	if row == nil {
		return nil, nil
	}
	item := &define.AgentCliItem{
		Id:                cast.ToInt(row[`id`]),
		Name:              cast.ToString(row[`name`]),
		Type:              cast.ToString(row[`type`]),
		SettingsPath:      cast.ToString(row[`settings_path`]),
		Config:            cast.ToString(row[`config`]),
		Enabled:           cast.ToInt(row[`enabled`]),
		ThinkingCollapsed: cast.ToInt(row[`thinking_collapsed`]),
		WebhookConfigId:   cast.ToInt(row[`webhook_config_id`]),
		CreatedAt:         cast.ToInt64(row[`created_at`]),
		UpdatedAt:         cast.ToInt64(row[`updated_at`]),
	}
	if item.Enabled != define.AgentCliEnabled {
		return nil, fmt.Errorf(`Agent CLI "%s" 已禁用`, item.Name)
	}
	return item, nil
}

// runClaudeCli 使用 Claude Code CLI 执行任务。
func runClaudeCli(cliItem *define.AgentCliItem, prompt string) *AgentCliResult {
	// 从 settings.json 读取模型和 API 配置
	model, baseURL, apiKey, err := parseClaudeSettings(cliItem.SettingsPath)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler-agent-cli] 解析 Claude settings 失败 %s`, err.Error())
		return &AgentCliResult{Content: fmt.Sprintf(`Claude 配置解析失败：%s`, err.Error()), Success: false}
	}

	cfg := p_claude.RunConfig{
		Prompt:       prompt,
		Model:        model,
		BaseURL:      baseURL,
		APIKey:       apiKey,
		WorkingDir:   getWorkingDir(),
		SettingsPath: cliItem.SettingsPath,
	}

	ctx, cancel := context.WithTimeout(context.Background(), agentCliTimeout)
	defer cancel()

	var resultBuilder strings.Builder
	var hasError bool

	sessionID, err := p_claude.RunClaudeStream(ctx, cfg, func(msg p_claude.StreamMessage) {
		// 收集助手文本输出
		switch msg.Type {
		case `assistant`:
			if text := extractClaudeAssistantText(msg); text != `` {
				resultBuilder.WriteString(text)
			}
		case `result`:
			// Claude 完成结果
			if result := cast.ToString(msg.Data[`result`]); result != `` {
				resultBuilder.WriteString(result)
			}
		}
		if msg.Type == `error` {
			hasError = true
			resultBuilder.WriteString(fmt.Sprintf(`\n[错误] %s`, msg.RawJSON))
		}
	})

	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler-agent-cli] Claude 执行失败 %s`, err.Error())
		return &AgentCliResult{
			Content:   fmt.Sprintf(`Claude 执行失败：%s\n\n%s`, err.Error(), resultBuilder.String()),
			Success:   false,
			SessionID: sessionID,
		}
	}

	content := resultBuilder.String()
	if content == `` {
		content = `任务已执行完成（无文本输出）`
	}
	if hasError {
		content = `任务执行过程中出现错误：\n` + content
	}

	gstool.FmtPrintlnLogTime(`[butler-agent-cli] Claude 执行完成 sessionID=%s`, sessionID)
	return &AgentCliResult{
		Content:   content,
		Success:   !hasError,
		SessionID: sessionID,
	}
}

// runCodexCli 使用 Codex CLI 执行任务。
func runCodexCli(cliItem *define.AgentCliItem, prompt string) *AgentCliResult {
	// 解析 Codex 专属配置 JSON
	var codexCfg define.CodexCliConfig
	if cliItem.Config != `` {
		if err := json.Unmarshal([]byte(cliItem.Config), &codexCfg); err != nil {
			gstool.FmtPrintlnLogTime(`[butler-agent-cli] 解析 Codex 配置 JSON 失败 %s`, err.Error())
			return &AgentCliResult{Content: fmt.Sprintf(`Codex 配置解析失败：%s`, err.Error()), Success: false}
		}
	}

	model := codexCfg.Model
	if model == `` && len(codexCfg.Models) > 0 {
		model = codexCfg.Models[0]
	}
	sandboxMode := codexCfg.SandboxMode
	if sandboxMode == `` {
		sandboxMode = define.CodexCliDefaultSandboxMode
	}

	cfg := p_codex.RunConfig{
		Prompt:      prompt,
		Model:       model,
		APIKey:      codexCfg.ApiKey,
		BaseURL:     codexCfg.BaseURL,
		WorkingDir:  getWorkingDir(),
		SandboxMode: sandboxMode,
	}

	ctx, cancel := context.WithTimeout(context.Background(), agentCliTimeout)
	defer cancel()

	var resultBuilder strings.Builder
	var hasError bool

	sessionID, err := p_codex.RunCodexStream(ctx, cfg, func(msg p_codex.StreamMessage) {
		// 收集助手文本输出
		switch msg.Type {
		case `item.completed`:
			if msg.ItemType == `agent_message` {
				if text := extractCodexItemText(msg); text != `` {
					resultBuilder.WriteString(text)
				}
			}
		}
		if msg.Type == `error` || msg.Type == `turn.failed` {
			hasError = true
			resultBuilder.WriteString(fmt.Sprintf(`\n[错误] %s`, msg.RawJSON))
		}
	})

	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler-agent-cli] Codex 执行失败 %s`, err.Error())
		return &AgentCliResult{
			Content:   fmt.Sprintf(`Codex 执行失败：%s\n\n%s`, err.Error(), resultBuilder.String()),
			Success:   false,
			SessionID: sessionID,
		}
	}

	content := resultBuilder.String()
	if content == `` {
		content = `任务已执行完成（无文本输出）`
	}
	if hasError {
		content = `任务执行过程中出现错误：\n` + content
	}

	gstool.FmtPrintlnLogTime(`[butler-agent-cli] Codex 执行完成 sessionID=%s`, sessionID)
	return &AgentCliResult{
		Content:   content,
		Success:   !hasError,
		SessionID: sessionID,
	}
}

// parseClaudeSettings 从 Claude Code 的 settings.json 中提取模型、baseURL、apiKey。
func parseClaudeSettings(settingsPath string) (model, baseURL, apiKey string, err error) {
	if settingsPath == `` {
		return ``, ``, ``, fmt.Errorf(`settings 路径为空`)
	}
	content, err := gstool.FileGetContent(settingsPath)
	if err != nil {
		return ``, ``, ``, fmt.Errorf(`读取 settings.json 失败: %w`, err)
	}
	var settings map[string]any
	if err := json.Unmarshal([]byte(content), &settings); err != nil {
		return ``, ``, ``, fmt.Errorf(`解析 settings.json 失败: %w`, err)
	}
	// 提取模型
	if env, ok := settings[`env`].(map[string]any); ok {
		model = cast.ToString(env[`ANTHROPIC_MODEL`])
		baseURL = cast.ToString(env[`ANTHROPIC_BASE_URL`])
		apiKey = cast.ToString(env[`ANTHROPIC_API_KEY`])
	}
	// 优先使用顶层配置
	if m := cast.ToString(settings[`model`]); m != `` {
		model = m
	}
	return model, baseURL, apiKey, nil
}

// extractClaudeAssistantText 从 Claude StreamMessage 中提取助手文本。
func extractClaudeAssistantText(msg p_claude.StreamMessage) string {
	if msg.Data == nil {
		return ``
	}
	// assistant 消息的 content 可能是字符串或数组
	content, ok := msg.Data[`content`]
	if !ok {
		return ``
	}
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var sb strings.Builder
		for _, item := range v {
			if itemMap, ok := item.(map[string]any); ok {
				if cast.ToString(itemMap[`type`]) == `text` {
					sb.WriteString(cast.ToString(itemMap[`text`]))
				}
			}
		}
		return sb.String()
	default:
		return cast.ToString(v)
	}
}

// extractCodexItemText 从 Codex StreamMessage 的 item 中提取文本。
func extractCodexItemText(msg p_codex.StreamMessage) string {
	if msg.Data == nil {
		return ``
	}
	item, ok := msg.Data[`item`].(map[string]any)
	if !ok {
		return ``
	}
	// agent_message 的 content 为字符串数组
	contentList, ok := item[`content`].([]any)
	if !ok {
		// 尝试直接作为字符串
		if text := cast.ToString(item[`content`]); text != `` {
			return text
		}
		return ``
	}
	var sb strings.Builder
	for _, c := range contentList {
		sb.WriteString(cast.ToString(c))
	}
	return sb.String()
}
