package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
	default:
		return fmt.Sprintf(`未知工具：%s`, name)
	}
}

// execFileRead 读取文件内容。
func execFileRead(path string) string {
	if path == `` {
		return `错误：文件路径不能为空`
	}
	data, err := os.ReadFile(path)
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
func execFileModify(path, search, replacement string) string {
	if path == `` {
		return `错误：文件路径不能为空`
	}
	if search == `` {
		return `错误：搜索文本不能为空`
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf(`读取文件失败：%s`, err.Error())
	}
	content := string(data)
	if !strings.Contains(content, search) {
		return `未找到匹配的文本`
	}
	newContent := strings.Replace(content, search, replacement, 1)
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return fmt.Sprintf(`写入文件失败：%s`, err.Error())
	}
	return `文件修改成功`
}

// execFileDelete 删除文件。
func execFileDelete(path string) string {
	if path == `` {
		return `错误：文件路径不能为空`
	}
	if err := os.Remove(path); err != nil {
		return fmt.Sprintf(`删除文件失败：%s`, err.Error())
	}
	return `文件删除成功`
}
