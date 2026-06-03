package stream

import (
	"bufio"
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
	reader := bufio.NewScanner(response.Body)
	//data 当前缓冲区的数据
	//atEOF 是否到达末尾
	//advance 告诉 Scanner 跳过多少字节（即已经处理了多少）
	//token Scan返回的数据
	var splitRE = regexp.MustCompile(h.Reges)
	reader.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		loc := splitRE.FindIndex(data)
		if loc != nil {
			return loc[1], data[:loc[1]], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	for reader.Scan() {
		resBytes := reader.Bytes()
		h.CallFunc(string(resBytes), nil)
		if h.FormatFunc != nil {
			*responseByte = append(*responseByte, h.FormatFunc(resBytes)...)
		} else {
			*responseByte = append(*responseByte, resBytes...)
		}
	}
}
