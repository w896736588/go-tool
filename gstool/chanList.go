package gstool

import (
	"container/list"
	"errors"
	"sync"
)

type ChanList struct {
	msgChan   chan any            // 用于接收消息的通道
	msgList   *list.List          // 用于存储消息的链表
	msgCall   func(msg any) error // 消息回调函数
	stopChan  chan struct{}       // 用于停止的通道
	pauseChan chan struct{}       // 用于暂停的通道
	mutex     sync.Mutex          // 用于保护链表的互斥锁
	wg        sync.WaitGroup      // 用于等待goroutine结束
	paused    bool                // 是否暂停触发回调
}

func (h *ChanList) Init(chanNum int, msgCall func(msg any) error) {
	h.msgChan = make(chan any, chanNum)
	h.msgList = list.New()
	h.msgCall = msgCall
	h.stopChan = make(chan struct{})
	h.pauseChan = make(chan struct{}, 1) // 用于控制暂停状态
	h.paused = false
	h.wg.Add(1)
	go h.processMessages()
}

func (h *ChanList) AddMsg(msg any) error {
	select {
	case h.msgChan <- msg:
		return nil
	default:
		return errors.New("通道已满")
	}
}

func (h *ChanList) processMessages() {
	defer h.wg.Done()
	for {
		select {
		case msg := <-h.msgChan:
			h.mutex.Lock()
			h.msgList.PushBack(msg)
			h.mutex.Unlock()

			if !h.isPaused() {
				h.processList()
			}
		case <-h.stopChan:
			return
		case <-h.pauseChan:
			h.mutex.Lock()
			h.paused = true
			h.mutex.Unlock()
		}
	}
}

func (h *ChanList) processList() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for e := h.msgList.Front(); e != nil; {
		next := e.Next()
		err := h.msgCall(e.Value)
		if err != nil {
			h.paused = true
			return
		}
		h.msgList.Remove(e)
		e = next
	}
}

func (h *ChanList) Pause() {
	h.pauseChan <- struct{}{}
}

func (h *ChanList) Active() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.paused {
		h.paused = false
	}
}

func (h *ChanList) isPaused() bool {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return h.paused
}

func (h *ChanList) Stop() {
	ChanClose(h.stopChan)
	h.wg.Wait()

	h.mutex.Lock()
	h.msgList.Init()
	h.mutex.Unlock()
}
