package worker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// skillsRoot 技能目录绝对路径，由外部在启动时设置，供路径解析使用。
var skillsRoot string

// dtoolBaseURL dtool API 基地址（如 http://localhost:17170），由外部在启动时设置。
var dtoolBaseURL string

// SetSkillsRoot 设置 skills 目录绝对路径，供文件工具解析相对路径时使用。
func SetSkillsRoot(root string) {
	skillsRoot = root
}

// SetDtoolBaseURL 设置 dtool API 基地址，供 http_call 工具拼接完整 URL。
func SetDtoolBaseURL(baseURL string) {
	dtoolBaseURL = baseURL
}

// ExecuteTool 执行指定的工具调用，返回执行结果文本。
func ExecuteTool(name string, argumentsJSON string) string {
	args := make(map[string]string)
	if err := json.Unmarshal([]byte(argumentsJSON), &args); err != nil {
		return fmt.Sprintf(`参数解析失败：%s`, err.Error())
	}
	switch name {
	case ToolFileRead:
		return execFileRead(args[`path`])
	case ToolFileWrite:
		return execFileWrite(args[`path`], args[`content`])
	case ToolFileModify:
		return execFileModify(args[`path`], args[`search`], args[`replacement`])
	case ToolFileDelete:
		return execFileDelete(args[`path`])
	case ToolHttpCall:
		return execHttpCall(args[`path`], args[`body`])
	case ToolRunScript:
		return execRunScript(args[`path`], args[`args`], args[`timeout`])
	case ToolAskUser:
		return execAskUser(args[`question`], args[`options`], args[`reason`])
	default:
		return fmt.Sprintf(`未知工具：%s`, name)
	}
}

// resolvePath 解析文件路径：如果是相对路径且直接读取失败，依次在 skills/dtool-butler/index/、skills/dtool-butler/scripts/ 和 skillsRoot 下查找。
// 优先级：直接路径 > skills/dtool-butler/index/ (索引文件) > skills/dtool-butler/scripts/ (自进化脚本) > skills/*/scripts/ (内置脚本)
func resolvePath(path string) (string, error) {
	if path == `` {
		return ``, fmt.Errorf(`文件路径不能为空`)
	}
	// 绝对路径直接使用
	if filepath.IsAbs(path) {
		return path, nil
	}
	// 先尝试原路径
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	if skillsRoot != `` {
		// 1. 优先在 skills/dtool-butler/index/ 下查找（索引文件：apis.md, scripts.md 等）
		indexCandidate := filepath.Join(skillsRoot, `dtool-butler`, `index`, path)
		if _, err := os.Stat(indexCandidate); err == nil {
			return indexCandidate, nil
		}
		// 2. 在 skills/dtool-butler/scripts/ 下查找（自进化生成的脚本）
		evolvedCandidate := filepath.Join(skillsRoot, `dtool-butler`, `scripts`, path)
		if _, err := os.Stat(evolvedCandidate); err == nil {
			return evolvedCandidate, nil
		}
		// 3. 回退：在 skillsRoot 下全面查找
		candidate := filepath.Join(skillsRoot, path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return path, nil
}

// execFileRead 读取文件内容。相对路径会自动在 skills 目录下查找。
func execFileRead(path string) string {
	resolved, err := resolvePath(path)
	if err != nil {
		return fmt.Sprintf(`读取文件失败：%s`, err.Error())
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return fmt.Sprintf(`读取文件失败：%s`, err.Error())
	}
	return string(data)
}

// execFileWrite 写入文件内容，自动创建父目录。
func execFileWrite(path, content string) string {
	if path == `` {
		return `错误：文件路径不能为空`
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Sprintf(`创建目录失败：%s`, err.Error())
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Sprintf(`写入文件失败：%s`, err.Error())
	}
	return `文件写入成功`
}

// execFileModify 查找并替换文件中的文本（仅替换第一个匹配项）。
// 相对路径会自动在 skills 目录下查找。
func execFileModify(path, search, replacement string) string {
	if path == `` {
		return `错误：文件路径不能为空`
	}
	if search == `` {
		return `错误：搜索文本不能为空`
	}
	resolved, err := resolvePath(path)
	if err != nil {
		return fmt.Sprintf(`读取文件失败：%s`, err.Error())
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return fmt.Sprintf(`读取文件失败：%s`, err.Error())
	}
	content := string(data)
	if !strings.Contains(content, search) {
		return `未找到匹配的文本`
	}
	newContent := strings.Replace(content, search, replacement, 1)
	if err := os.WriteFile(resolved, []byte(newContent), 0644); err != nil {
		return fmt.Sprintf(`写入文件失败：%s`, err.Error())
	}
	return `文件修改成功`
}

// execFileDelete 删除文件。相对路径会自动在 skills 目录下查找。
func execFileDelete(path string) string {
	if path == `` {
		return `错误：文件路径不能为空`
	}
	resolved, err := resolvePath(path)
	if err != nil {
		return fmt.Sprintf(`删除文件失败：%s`, err.Error())
	}
	if err := os.Remove(resolved); err != nil {
		return fmt.Sprintf(`删除文件失败：%s`, err.Error())
	}
	return `文件删除成功`
}

// execRunScript 执行本地 Python 脚本并返回 stdout+stderr。
// 脚本路径会自动在 skillsRoot 下查找。
func execRunScript(path, argsStr, timeoutStr string) string {
	if path == `` {
		return `错误：脚本路径不能为空`
	}
	// 解析路径
	resolved, err := resolvePath(path)
	if err != nil {
		return fmt.Sprintf(`脚本路径解析失败：%s`, err.Error())
	}
	// 检查文件存在
	if _, err := os.Stat(resolved); err != nil {
		return fmt.Sprintf(`脚本不存在：%s`, resolved)
	}
	// 构建命令参数
	cmdArgs := []string{resolved}
	if argsStr != `` {
		cmdArgs = append(cmdArgs, strings.Fields(argsStr)...)
	}
	cmd := exec.Command(`python`, cmdArgs...)
	output, err := cmd.CombinedOutput()
	result := string(output)
	if err != nil {
		return fmt.Sprintf(`脚本执行失败：%s\n输出：%s`, err.Error(), truncateForLog(result, 2000))
	}
	if len(result) > 3000 {
		result = result[:3000] + `\n...(输出已截断)`
	}
	return result
}

// execAskUser 向用户发起确认问题。
// 返回特殊格式的标记字符串，供 FC 循环检测并暂停等待用户回复。
func execAskUser(question, options, reason string) string {
	if question == `` {
		return `错误：确认问题不能为空`
	}
	result := map[string]string{
		`marker`:   AskUserMarker,
		`question`: question,
		`options`:  options,
		`reason`:   reason,
	}
	b, _ := json.Marshal(result)
	return string(b)
}

// execHttpCall 调用 dtool 的 HTTP API 接口。
// 自动拼接 dtooolBaseURL 与传入的 path，发起 POST 请求并返回响应文本。
func execHttpCall(path, body string) string {
	if path == `` {
		return `错误：API 路径不能为空`
	}
	if dtoolBaseURL == `` {
		return `错误：dtool API 基地址未配置，无法发起 HTTP 调用`
	}
	// 确保 path 以 / 开头
	if !strings.HasPrefix(path, `/`) {
		path = `/` + path
	}
	fullURL := strings.TrimRight(dtoolBaseURL, `/`) + path

	req, err := http.NewRequest(http.MethodPost, fullURL, strings.NewReader(body))
	if err != nil {
		return fmt.Sprintf(`创建 HTTP 请求失败：%s`, err.Error())
	}
	req.Header.Set(`Content-Type`, `application/json`)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf(`HTTP 请求失败：%s`, err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf(`读取响应失败：%s`, err.Error())
	}
	// 截断过长响应
	result := string(respBody)
	if len(result) > 3000 {
		result = result[:3000] + `\n...(响应已截断)`
	}
	return fmt.Sprintf(`HTTP %d %s → %s`, resp.StatusCode, fullURL, result)
}
