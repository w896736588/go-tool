package stream

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// BytsNum 按照字节数截取
type BytsNum struct {
	Num        int                 //按多少个字节截取一次
	CallFunc   func(string, error) //截取后的回调
	FormatFunc func([]byte) []byte //截取后的数据处理回调 这个结果将会存入最终返回的数据中，如果不设置就按照实际接收的计入最终结果
}

func (h *BytsNum) ReceiveSplit(response *http.Response, responseByte *[]byte) {
	reader := io.Reader(response.Body)
	buffer := make([]byte, h.Num)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			h.CallFunc(``, errors.New(fmt.Sprintf("Error reading response: %v", err)))
			break
		}
		h.CallFunc(string(buffer[:n]), nil)
		if h.FormatFunc != nil {
			*responseByte = append(*responseByte, h.FormatFunc(buffer[:n])...)
		} else {
			*responseByte = append(*responseByte, buffer[:n]...)
		}
	}
}
