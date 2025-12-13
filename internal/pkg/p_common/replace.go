package p_common

import (
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

// Replace 替换变量
func Replace(data string, replaceList map[string]string) string {
	//处理特殊情况
	for replaceKey, replaceVal := range replaceList {
		//取模
		matchSubList := gstool.RegexMatchSubString(data, replaceKey+`%(\d+)`)
		if len(matchSubList) >= 2 {
			data = gstool.SReplaces(data, map[string]string{
				matchSubList[0]: cast.ToString(cast.ToInt64(replaceVal) % cast.ToInt64(matchSubList[1])),
			})
		}
	}
	data = gstool.SReplaces(data, replaceList)
	return data
}
