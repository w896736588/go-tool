package stream

import (
	"net/http"
	"regexp"
)

// Reges 按照正则截取
type Reges struct {
	Reges      string              //按正则进行分割如果是ascii 那么应该是\x00-\x1F
	CallFunc   func(string, error) //截取后的回调
	FormatFunc func([]byte) []byte //截取后的数据处理回调 这个结果将会存入最终返回的数据中，如果不设置就按照实际接收的计入最终结果
}

func (h *Reges) ReceiveSplit(response *http.Response, responseByte *[]byte) {
	if h.CallFunc == nil {
		return
	}
	splitRE := regexp.MustCompile(h.Reges)
	readAndSplit(response.Body, h.CallFunc, h.FormatFunc, responseByte,
		func(data []byte) int {
			if loc := splitRE.FindIndex(data); loc != nil {
				return loc[1]
			}
			return -1
		},
	)
}
