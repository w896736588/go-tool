//go:build windows

package gstool

import (
	"syscall"
	"time"
)

// FileEditCreateTime 修改文件创建时间、最后访问时间、最后修改时间
func FileEditCreateTime(filePath string, ctime, atime, mtime time.Time) error {
	pathPtr, pathPtrErr := syscall.UTF16PtrFromString(filePath)
	if pathPtrErr != nil {
		return pathPtrErr
	}
	handle, handleErr := syscall.CreateFile(pathPtr, syscall.FILE_WRITE_ATTRIBUTES, syscall.FILE_SHARE_WRITE, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if handleErr != nil {
		return handleErr
	}
	defer func(fd syscall.Handle) {
		closeErr := syscall.Close(fd)
		if closeErr != nil {
			FmtPrintlnLog(`close error：%s`, closeErr.Error())
		}
	}(handle)
	a := syscall.NsecToFiletime(syscall.TimespecToNsec(syscall.NsecToTimespec(atime.UnixNano())))
	c := syscall.NsecToFiletime(syscall.TimespecToNsec(syscall.NsecToTimespec(ctime.UnixNano())))
	m := syscall.NsecToFiletime(syscall.TimespecToNsec(syscall.NsecToTimespec(mtime.UnixNano())))
	return syscall.SetFileTime(handle, &c, &a, &m)
}
