package gstool

import (
	"errors"
	"os"
	"path/filepath"
)

// GetRootPath 获取项目根目录 注意 根目录下面需要放一个 go.mod文件
// 编译后运行需要使用 wd, _ = os.Executable() 获取wd
// go run 运行需要使用 _, wd, _, _ = runtime.Caller(0) 获取wd
func GetRootPath(wd string) (string, error) {
	// 找到项目根目录
	rootDir := findRootDir(wd)
	if rootDir == `` {
		return ``, errors.New(`failed to find project root directory`)
	}
	return rootDir, nil
}

// 找到项目根目录
func findRootDir(dir string) string {
	// 如果已经到达根目录，则返回空字符串
	if dir == filepath.Dir(dir) {
		return ``
	}

	// 判断是否存在 go.mod 文件
	modFile := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(modFile); err == nil {
		return dir
	}

	// 向上一级目录查找
	return findRootDir(filepath.Dir(dir))
}

// DirPathExists 目录是否存在
func DirPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// DirCreatePath 创建目录 支持多级目录创建
func DirCreatePath(path string) error {
	exists, dirError := DirPathExists(path)
	if dirError != nil {
		return dirError
	}
	if !exists {
		//创建目录
		createDirErr := os.MkdirAll(path, os.ModePerm)
		if createDirErr != nil {
			return createDirErr
		}
	}
	return nil
}

// DirWalk 遍历文件夹下面所有文件夹和文件
func DirWalk(rootDir string, back func(path string, info os.FileInfo, err error)) error {
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		back(path, info, err)
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

// DirWalkFirstLevel 遍历文件夹下面第一级所有目录和文件
func DirWalkFirstLevel(rootDir string, back func(info os.DirEntry)) error {
	entityList, err := os.ReadDir(rootDir)
	if err != nil {
		return err
	}
	for _, entry := range entityList {
		back(entry)
	}
	return nil
}

// DirIsGit 判断目录是否是git仓库
func DirIsGit(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		return false
	}
	return info.IsDir()
}

func DirPathFormatToWindows(dirPath string) string {
	return filepath.FromSlash(filepath.ToSlash(dirPath))
}

func DirIsEmpty(dirPath string) (bool, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

// DirRemoveEmpty 删除文件夹（仅当为空时）
func DirRemoveEmpty(dirPath string) error {
	return os.Remove(dirPath)
}
