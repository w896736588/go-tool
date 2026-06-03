package gstool

import (
	"crypto/md5"
	"fmt"
	"github.com/spf13/cast"
	"io"
)

// Md5 md加密
func Md5(str string) string {
	w := md5.New()
	_, err := io.WriteString(w, str)
	if err != nil {
		return ""
	}
	//将str写入到w中
	return fmt.Sprintf("%x", w.Sum(nil))
}

// Md5Slice md加密
func Md5Slice(mapData ...any) string {
	str := ``
	for _, value := range mapData {
		str += cast.ToString(value)
	}
	return Md5(str)
}
