package lib_tool

import (
	"crypto/md5"
	"fmt"
	"io"
)

// Md5
// @auth frog
// @date 2023-04-13 09:38:37
func Md5(str string) string {
	w := md5.New()
	_, err := io.WriteString(w, str)
	if err != nil {
		return ""
	}
	//将str写入到w中
	return fmt.Sprintf("%x", w.Sum(nil))
}
