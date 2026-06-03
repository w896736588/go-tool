package gsgin

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var SseList map[string]*Sse
var lock sync.RWMutex

func init() {
	SseList = make(map[string]*Sse)
}

type Event struct {
	EventId string
	Msg     string
}

type Sse struct {
	ClientId    string           //链接ID
	StopC       chan int         //控制gin接口是否退出
	C           *gin.Context     //c gin
	ChanList    *gstool.ChanList //消息通道
	lastEventId int64            //最后发送的eventId
}

func SseRegister(clientId string, stopC chan int, c *gin.Context) *Sse {
	if _, ok := SseList[clientId]; ok {
		sse := SseList[clientId]
		sse.StopC = stopC
		sse.C = c
		sse.ChanList = &gstool.ChanList{}
		sse.lastEventId = time.Now().UnixMilli()
		sse.ChanList.Init(10000, func(msg any) error {
			return sse.Send(msg.(*Event))
		})
		return sse
	}
	chanList := &gstool.ChanList{}
	sse := &Sse{
		ClientId:    clientId,
		StopC:       stopC,
		C:           c,
		ChanList:    chanList,
		lastEventId: time.Now().UnixMilli(),
	}
	chanList.Init(10000, func(msg any) error {
		return sse.Send(msg.(*Event))
	})
	lock.Lock()
	defer lock.Unlock()
	SseList[clientId] = sse
	return sse
}

func SseStatus() []string {
	lock.Lock()
	defer lock.Unlock()
	msgList := make([]string, 0)
	for _, v := range SseList {
		msgList = append(msgList, fmt.Sprintf(`ClientId:%s`, v.ClientId))
	}
	return msgList
}

func (h *Sse) Send(e *Event) (returnErr error) {
	if h == nil {
		return nil
	}
	defer func() {
		if err := recover(); err != nil {
			returnErr = gstool.Error(`%s 发送失败 %s`, h.ClientId, err)
		}
	}()
	sendMsg := fmt.Sprintf("id: %s\ndata: %s\n\n", e.EventId, e.Msg)
	_, err := h.C.Writer.Write([]byte(sendMsg))
	if err != nil {
		returnErr = gstool.Error(`%s 写入失败 %s %s`, h.ClientId, sendMsg, err.Error())
		return
	} else {
		h.C.Writer.Flush()
	}
	return
}

func (h *Sse) SendToChan(msg string) error {
	if h == nil {
		return nil
	}
	h.lastEventId++
	return h.ChanList.AddMsg(&Event{
		EventId: fmt.Sprintf(`%d`, h.lastEventId),
		Msg:     msg,
	})
}

func (h *Sse) Pause() {
	if h == nil {
		return
	}
	h.ChanList.Pause()
}

func (h *Sse) UnRegister() {
	if h == nil {
		return
	}
	h.CleanMsg()
	SseDelete(h.ClientId)
}

func SseDelete(clientId string) {
	lock.Lock()
	defer lock.Unlock()
	delete(SseList, clientId)
}

func (h *Sse) CleanMsg() {
	if h == nil {
		return
	}
	h.Pause()
	h.ChanList.Stop()
	chanList := &gstool.ChanList{}
	chanList.Init(10000, func(msg any) error {
		return h.Send(msg.(*Event))
	})
	h.ChanList = chanList
}

func SseGetByClientId(clientId string) *Sse {
	lock.Lock()
	defer lock.Unlock()
	if sse, ok := SseList[clientId]; ok {
		return sse
	}
	return nil
}
