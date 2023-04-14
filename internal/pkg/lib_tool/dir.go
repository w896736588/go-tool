package lib_tool

import (
	"os"
	"path"
	"runtime"
)

//获取调用处往上的目录
func DirUpNum(upNum int) string {
	_, filename, _, _ := runtime.Caller(0)
	for i := 0; i < upNum; i++ {
		filename = path.Dir(filename)
	}
	return filename
}

//目录是否存在
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

//创建目录
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
