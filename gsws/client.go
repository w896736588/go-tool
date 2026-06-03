package gsws

import (
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cast"
)

type GsWsClient struct {
	Host       string
	Port       string
	Uri        string
	RecMsgChan chan string
	CallFunc   func(string) string
}

func (h *GsWsClient) Start() error {
	h.RecMsgChan = make(chan string)
	u := url.URL{Scheme: "ws", Host: h.Host + `:` + h.Port, Path: `/` + h.Uri}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	go func() {
		defer close(h.RecMsgChan)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				gstool.FmtPrintlnLog(`接收消息失败 %s`, err.Error())
				return
			}
			h.RecMsgChan <- cast.ToString(message)
		}
	}()

	//ticker := time.NewTicker(time.Second)
	//defer ticker.Stop()
	for {
		select {
		// if the goroutine is done , all are out
		case msg := <-h.RecMsgChan: //正常消息处理
			h.CallFunc(msg)
		case t := <-time.Tick(time.Second): //主动发起心跳
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				gstool.FmtPrintlnLog(`发送心跳失败 %s`, err.Error())
				break
			}
			return nil
		}
	}
}

// SetCallFunc 设置回调
func (h *GsWsClient) SetCallFunc(callFunc func(string) string) {
	h.CallFunc = callFunc
}
