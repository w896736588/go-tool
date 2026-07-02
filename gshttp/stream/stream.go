package stream

import (
	"errors"
	"io"
)

const maxPendingSize = 2 << 20 // 2MB 缓冲区上限，防止内存无限增长

// flushPending 将 pending 中剩余数据作为最后一块回调出去
func flushPending(callFunc func(string, error), formatFunc func([]byte) []byte, pending *[]byte, responseByte *[]byte) {
	if len(*pending) == 0 {
		return
	}
	callFunc(string(*pending), nil)
	if formatFunc != nil {
		*responseByte = append(*responseByte, formatFunc(*pending)...)
	} else {
		*responseByte = append(*responseByte, *pending...)
	}
	*pending = nil
}

// tokenMatcher 在 data 上查找分隔符，返回匹配结束位置；未找到返回 -1
type tokenMatcher func(data []byte) int

// readAndSplit 通用流读取与分割循环
func readAndSplit(
	reader io.Reader,
	callFunc func(string, error),
	formatFunc func([]byte) []byte,
	responseByte *[]byte,
	matcher tokenMatcher,
) {
	buf := make([]byte, 4096)
	var pending []byte
	var scanFrom int

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			pending = append(pending, buf[:n]...)
		}
		// 从上次位置继续查找分隔符并切分
		for {
			if scanFrom >= len(pending) {
				break
			}
			end := matcher(pending[scanFrom:])
			if end < 0 {
				break
			}
			absEnd := scanFrom + end
			token := pending[:absEnd]
			pending = pending[absEnd:]
			scanFrom = 0 // 切分后 pending 已变化，重置扫描偏移
			callFunc(string(token), nil)
			if formatFunc != nil {
				*responseByte = append(*responseByte, formatFunc(token)...)
			} else {
				*responseByte = append(*responseByte, token...)
			}
		}
		if err != nil {
			flushPending(callFunc, formatFunc, &pending, responseByte)
			if !errors.Is(err, io.EOF) {
				callFunc(``, errors.New("Error reading response: "+err.Error()))
			}
			break
		}
		// pending 过大且仍无匹配时，强制 flush 避免内存膨胀
		if len(pending) > maxPendingSize {
			callFunc(``, errors.New("pending buffer exceeded max size without delimiter"))
			flushPending(callFunc, formatFunc, &pending, responseByte)
			break
		}
	}
}
