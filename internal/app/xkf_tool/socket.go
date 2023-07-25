package xkf_tool

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	"gitee.com/Sxiaobai/gs/gsws"
	"time"
)

type ServerConn struct {
	wsConn *gsws.WsConn
}

// InitXkfSocket 初始化socket
func InitXkfSocket() {
	wsConfig := gsws.Config{
		Host:             `0.0.0.0`,
		Port:             ConfigViper.GetString(`run.wsXkfPort`),
		Uri:              `/conn`,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		CheckOrigin:      false,
		MaxLiveTime:      60,
		MaxWriteWaitTime: 60,
		MaxReadWaitTime:  60,
		MaxMessageSize:   1024,
		MaxClient:        100,
	}
	wss := gsws.Server{}
	wss.SetConfig(wsConfig)
	//生成客户端ID
	wss.SetGClientIdFunc(func() string {
		return fmt.Sprintf(`%d_%d`, time.Now().Unix(), gstool.RandNumber(10000, 99999))
	})
	//接收到新连接
	wss.SetNewConnFunc(func(conn *gsws.WsConn, queryParams map[string]any) {
		serverConn := &ServerConn{
			wsConn: conn,
		}
		conn.SetReceiveFunc(serverConn.ReceiveXkfMsg)
		gstool.FmtPrintlnLog(`增加新的连接 %s`, conn.GetClientId())
	})
	wss.SetConnCloseFunc(func(clientId string) {
		gstool.FmtPrintlnLog(`断开的连接 %s`, clientId)
	})
	err := wss.Start()

	if err != nil {
		gstool.FmtPrintlnLog(`启动失败 %s`, err.Error())
	}
}

// InitWkSocket 初始化socket
func InitWkSocket() {
	wsConfig := gsws.Config{
		Host:             `0.0.0.0`,
		Port:             ConfigViper.GetString(`run.wsWkPort`),
		Uri:              `/conn`,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		CheckOrigin:      false,
		MaxLiveTime:      60,
		MaxWriteWaitTime: 60,
		MaxReadWaitTime:  60,
		MaxMessageSize:   1024,
		MaxClient:        100,
	}
	wss := gsws.Server{}
	wss.SetConfig(wsConfig)
	//生成客户端ID
	wss.SetGClientIdFunc(func() string {
		return fmt.Sprintf(`%d_%d`, time.Now().Unix(), gstool.RandNumber(10000, 99999))
	})
	//接收到新连接
	wss.SetNewConnFunc(func(conn *gsws.WsConn, queryParams map[string]any) {
		serverConn := &ServerConn{
			wsConn: conn,
		}
		conn.SetReceiveFunc(serverConn.ReceiveWkMsg)
		gstool.FmtPrintlnLog(`增加新的连接 %s`, conn.GetClientId())
	})
	wss.SetConnCloseFunc(func(clientId string) {
		gstool.FmtPrintlnLog(`断开的连接 %s`, clientId)
	})
	err := wss.Start()

	if err != nil {
		gstool.FmtPrintlnLog(`启动失败 %s`, err.Error())
	}
}

func (h *ServerConn) ReceiveXkfMsg(message string) {
	if message == `ping` {
		err := h.wsConn.SendMessage(`pong`)
		if err != nil {
			gstool.FmtPrintlnLog(`回复消息错误 %s`, err.Error())
			return
		}
		return
	}

}

func (h *ServerConn) ReceiveWkMsg(message string) {
	if message == `ping` {
		err := h.wsConn.SendMessage(`pong`)
		if err != nil {
			gstool.FmtPrintlnLog(`回复消息错误 %s`, err.Error())
			return
		}
		return
	}
}
