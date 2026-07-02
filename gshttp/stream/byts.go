package stream

import (
	"bytes"
	"net/http"
)

// Byts 按照字符串截取
type Byts struct {
	Byts       []byte              //按字符串进行分割如果是ascii 那么应该是\x00-\x1F
	CallFunc   func(string, error) //截取后的回调
	FormatFunc func([]byte) []byte //截取后的数据处理回调 这个结果将会存入最终返回的数据中，如果不设置就按照实际接收的计入最终结果
}

func (h *Byts) ReceiveSplit(response *http.Response, responseByte *[]byte) {
	if h.CallFunc == nil {
		return
	}
	readAndSplit(response.Body, h.CallFunc, h.FormatFunc, responseByte,
		func(data []byte) int {
			if idx := bytes.Index(data, h.Byts); idx >= 0 {
				return idx + len(h.Byts)
			}
			return -1
		},
	)
}
