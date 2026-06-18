package index

import (
	"dev_tool/internal/app/dtool/define"
	"os"
	"path/filepath"

	"github.com/w896736588/go-tool/gstool"
)

// 索引文件名常量
const (
	CapabilitiesFileName = `capabilities.md` // 总能力清单
	ScriptsFileName      = `scripts.md`      // 脚本工具索引
	ApisFileName         = `apis.md`         // dtool HTTP 接口索引
)

// ResolveIndexPath 解析索引文档目录路径。
// 优先使用 config.IndexDocPath，为空时回落到 {memoryDbPath}/butler/index/。
func ResolveIndexPath(config *define.ButlerConfigItem, env *define.ButlerEnv) string {
	if config.IndexDocPath != `` {
		return config.IndexDocPath
	}
	return filepath.Join(env.MemoryDbPath, `butler`, `index`)
}

// EnsureIndexDir 确保索引目录存在。
func EnsureIndexDir(indexPath string) error {
	return gstool.DirCreatePath(indexPath)
}

// ReadIndexFile 读取索引文件内容，文件不存在时返回空字符串。
func ReadIndexFile(indexPath, fileName string) string {
	filePath := filepath.Join(indexPath, fileName)
	content, err := gstool.FileGetContent(filePath)
	if err != nil {
		return ``
	}
	return content
}

// WriteIndexFile 写入索引文件内容，自动创建父目录。
func WriteIndexFile(indexPath, fileName, content string) error {
	filePath := filepath.Join(indexPath, fileName)
	if err := gstool.DirCreatePath(filepath.Dir(filePath)); err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

// IndexExists 检查索引文件是否存在且非空。
func IndexExists(indexPath, fileName string) bool {
	filePath := filepath.Join(indexPath, fileName)
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return info.Size() > 0
}
