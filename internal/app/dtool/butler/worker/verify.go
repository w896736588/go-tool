package worker

import (
	"fmt"
	"os"
	"strings"
)

// VerifyResult 验收结果。
type VerifyResult struct {
	Passed  bool   // 是否通过验收
	Message string // 验收说明
}

// VerifyFileExists 验证文件是否存在。
func VerifyFileExists(path string) VerifyResult {
	if path == `` {
		return VerifyResult{Passed: false, Message: `文件路径为空`}
	}
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return VerifyResult{Passed: false, Message: fmt.Sprintf(`文件不存在：%s`, path)}
		}
		return VerifyResult{Passed: false, Message: fmt.Sprintf(`检查文件失败：%s`, err.Error())}
	}
	return VerifyResult{Passed: true, Message: fmt.Sprintf(`文件存在：%s`, path)}
}

// VerifyFileContains 验证文件是否包含指定内容。
func VerifyFileContains(path, expectedContent string) VerifyResult {
	data, err := os.ReadFile(path)
	if err != nil {
		return VerifyResult{Passed: false, Message: fmt.Sprintf(`读取文件失败：%s`, err.Error())}
	}
	if !strings.Contains(string(data), expectedContent) {
		return VerifyResult{Passed: false, Message: `文件不包含预期内容`}
	}
	return VerifyResult{Passed: true, Message: `文件内容验证通过`}
}

// VerifyFileNotExists 验证文件已被删除。
func VerifyFileNotExists(path string) VerifyResult {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return VerifyResult{Passed: true, Message: fmt.Sprintf(`文件已删除：%s`, path)}
		}
		return VerifyResult{Passed: false, Message: fmt.Sprintf(`检查文件失败：%s`, err.Error())}
	}
	return VerifyResult{Passed: false, Message: fmt.Sprintf(`文件仍存在：%s`, path)}
}
