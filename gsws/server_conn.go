package gsws

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cast"
)

// WsConn 客户端连接配置
type WsConn struct {
	conn      *websocket.Conn //连接
	readChan  chan string     //接收管道
	writeChan chan string     //输出管道

	mutex     sync.Mutex // 避免重复关闭管道,加锁处理
	isClosed  bool
	closeChan chan byte // 关闭通知
	clientId  string    //客户端ID

	receiveMsgFunc func(item string) //读到消息以后的处理
	serverHandle   *Server

	lastHeartTime int64 //上一次心跳维持的时间
}

// Message 客户端读写消息
type Message struct {
	WebsocketMessageType int64 //websocket的消息类型
	Data                 string
}

// GsMessageItem 消息体
type GsMessageItem struct {
	Type int    `json:"type"`
	Msg  string `json:"msg"`
}

// IsConnected 客户端是否在连接中
func (h *WsConn) IsConnected() bool {
	if time.Now().Unix()-h.lastHeartTime < h.serverHandle.config.MaxLiveTime {
		return false
	} else {
		return true
	}
}

// 设置最大读取等待时间 每次读取到消息以后都要再次设置
func (h *WsConn) setReadDeadline() {
	err := h.conn.SetReadDeadline(time.Now().Add(h.serverHandle.config.MaxReadWaitTime))
	if err != nil {
		gstool.FmtPrintlnLog(`设置最大读取时间错误 %s`, err.Error())
		return
	}
}

// 设置写超时时间
func (h *WsConn) setWriteDeadline() {
	err := h.conn.SetWriteDeadline(time.Now().Add(h.serverHandle.config.MaxWriteWaitTime))
	if err != nil {
		gstool.FmtPrintlnLog(`设置最大读取时间错误 %s`, err.Error())
		return
	}
}

// 处理队列中的消息
func (h *WsConn) doBusinessLoop() {
	// 获取到消息队列中的消息，处理完成后，发送消息给客户端
	for {
		msg, err := h.read()
		if err != nil {
			gstool.FmtPrintlnLog(`读取消息失败 %s`, err.Error())
			break
		}
		gstool.FmtPrintlnLog(`接收到消息 %s`, msg)
		//执行业务
		h.receiveMsgFunc(msg)

	}
}

// 读取消息队列中的消息
func (h *WsConn) read() (string, error) {
	select {
	case msg := <-h.readChan:
		// 获取到消息队列中的消息
		return msg, nil
	case <-h.closeChan:
		return ``, errors.New(`连接已经关闭`)
	}

}

// SendMessage 写入消息到队列中
func (h *WsConn) SendMessage(backMsg string) error {
	select {
	case h.writeChan <- backMsg:
	case <-h.closeChan:
		return errors.New(`连接已经关闭`)
	}
	return nil
}

// 处理消息队列中的消息
func (h *WsConn) wsReadLoop() {
	for {
		h.setReadDeadline()
		h.setWriteDeadline()
		// 读一个message
		messageType, receiveData, err := h.conn.ReadMessage()
		if err != nil {
			//错误是否为某些类型
			//websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure)
			gstool.FmtPrintlnLog(`消息读取出现错误 %d %s`, messageType, err.Error())
			h.close()
			return
		}

		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			gstool.FmtPrintlnLog(`断开连接消息 结束 %d %s`, messageType, err.Error())
			h.close()
			return
		}
		receiveMsg := cast.ToString(receiveData)
		//ping消息
		if receiveMsg == `ping` {
			h.lastHeartTime = time.Now().Unix()
			h.writeChan <- `pong`
			return
		}

		// 放入请求队列,消息入栈
		select {
		case h.readChan <- receiveMsg:
		case <-h.closeChan:
			gstool.FmtPrintlnLog(`收到中断信号 结束循环read`)
			return
		}
	}
}

// 发送消息给客户端
func (h *WsConn) wsWriteLoop() {
	for {
		select {
		case msg := <-h.writeChan:
			if err := h.conn.WriteMessage(websocket.TextMessage, []byte(gstool.JsonEncode(msg))); err != nil {
				gstool.FmtPrintlnLog(`发送消息给客户端发生错误 %s`, err.Error())
				h.close()
				return
			}
		case <-h.closeChan:
			// 获取到关闭通知
			gstool.FmtPrintlnLog(`收到关闭通知 结束写入循环`)
			return
		}
	}
}

// 关闭连接
func (h *WsConn) close() {
	h.conn.Close()
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if h.isClosed == false {
		h.isClosed = true
		h.serverHandle.removeCliConn(h)
		close(h.closeChan)
	}
}

// GetClientId 返回客户端ID
func (h *WsConn) GetClientId() string {
	return h.clientId
}
