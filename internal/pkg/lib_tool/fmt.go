package lib_tool

import "fmt"

// FmtPrintlnLog 标准输出日志
// @auth frog
// @date 2023-04-13 10:50:32
func FmtPrintlnLog(msg string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(msg, args...))
}
