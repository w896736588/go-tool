package gstool

import "fmt"

// FmtPrintlnLog 标准输出日志
func FmtPrintlnLog(msg string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(msg, args...))
}

// FmtPrintlnLogTime 标准输出日志 增加时间
func FmtPrintlnLogTime(msg string, args ...interface{}) {
	fmt.Println(TimeNowUnixToString(`Y-m-d H:i:s`) + ` ` + fmt.Sprintf(msg, args...))
}

// FmtLogTime 增加时间格式化字符串
func FmtLogTime(msg string, args ...interface{}) string {
	return TimeNowUnixToString(`Y-m-d H:i:s`) + ` ` + fmt.Sprintf(msg, args...)
}
