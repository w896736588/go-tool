package gstool

import (
	"errors"
	"sync"
	"time"
)

type ChanTask struct {
	Chan     chan interface{}
	CallFunc func(interface{})
	isStop   bool
	wg       sync.WaitGroup // 用于等待所有任务完成
}

// ChanCreate 创建一个并发任务
// chanMaxNum 通道最多可以存储多少个消息
// maxDoNum 多少个协程处理消息，并发数
func ChanCreate(chanMaxNum, maxDoNum int, callFunc func(interface{})) *ChanTask {
	chanStruct := ChanTask{
		Chan:     make(chan interface{}, chanMaxNum),
		CallFunc: callFunc,
	}
	for i := 0; i < maxDoNum; i++ {
		go chanStruct.Do()
	}
	return &chanStruct
}

// Add 加入消息
func (h *ChanTask) Add(msg interface{}, timeOut time.Duration) error {
	if h.isStop {
		return errors.New("通道已经关闭")
	}
	select {
	case h.Chan <- msg:
		return nil
	case <-time.After(timeOut): // 超时时间可以根据需要调整
		return errors.New("添加消息超时")
	}
}

// Do 消费
func (h *ChanTask) Do() {
	for {
		select {
		case msg, ok := <-h.Chan:
			if !ok {
				return // 通道关闭，退出协程
			}
			h.wg.Add(1) // 增加等待组计数
			h.call(msg)
		}
	}
}

func (h *ChanTask) call(msg any) {
	defer h.wg.Done() // 任务完成，减少等待组计数
	h.CallFunc(msg)
}

// Stop 关闭
func (h *ChanTask) Stop() {
	h.isStop = true //设置停止 停止往里面写入
	h.wg.Wait()     // 等待所有任务完成
	close(h.Chan)   //关闭通道
}
