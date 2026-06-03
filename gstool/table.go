package gstool

import (
	"fmt"
)
import "github.com/spf13/cast"

// TableMod 取模后的值格式化
// 例如：TableMod(2,2) 将2格式化为第二个长度02  TableMod(2,1) 将返回2  TableMod(10 , 3) 将返回010
func TableMod(mainKey, length int) string {
	return fmt.Sprintf(`%0`+cast.ToString(length)+`d`, mainKey)
}

// TableModAscii 根据字符串取模进行分表 将字符串转为ascii码并累加后进行取模 仅支持ascii字符串
// tableNum 表示分表数
// TableModAscii(`aaaaab`, 10) 输出00
func TableModAscii(mainKey string, tableNum int) string {
	if tableNum == 1 { //分一个表无意义
		return ``
	}
	asciiInt := StringToAsciiInt(mainKey)
	return TableMod(asciiInt%tableNum, len(cast.ToString(tableNum)))
}

// TableModAsciiNoZero 根据字符串取模进行分表 将字符串转为ascii码并累加后进行取模 仅支持ascii字符串
// tableNum 表示分表数
// TableModAsciiNoZero(`aaaaab`, 10) 输出0
func TableModAsciiNoZero(mainKey string, tableNum int) string {
	if tableNum == 1 { //分一个表无意义
		return ``
	}
	asciiInt := StringToAsciiInt(mainKey)
	return cast.ToString(asciiInt % tableNum)
}
