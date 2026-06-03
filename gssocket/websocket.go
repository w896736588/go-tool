package gssocket

import (
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

type Server struct {
	Host          string
	Uri           string                        //接口
	ReceMsgFunc   func(string, string)          //收消息函数
	WriteMsgFunc  func(string)                  //发送消息函数
	ConnectFunc   func(string, *websocket.Conn) //建立链接后回调
	CloseFunc     func(string)                  //关闭链接回调
	GetClientFunc func(r *http.Request) string
	AllowOrigin   bool
	ClientConnMap *gstool.HighMap //clientId与conn之间的映射
	upgrader      websocket.Upgrader
}

func (h *Server) Start() {
	h.ClientConnMap = gstool.HighMapCreate(10)
	h.upgrader = websocket.Upgrader{} // use default options
	if h.AllowOrigin {
		h.upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}
	http.HandleFunc(h.Uri, h.newConn)
	err := http.ListenAndServe(h.Host, nil)
	if err != nil {
		gstool.FmtPrintlnLog(`初始化socket 失败 %s`, err.Error())
	}
}

func (h *Server) WriteMessage(clientId, backMsg string) error {
	conn, boolConn := h.ClientConnMap.Get(clientId)
	if !boolConn {
		return errors.New(`不存在的链接`)
	}

	err := conn.(*websocket.Conn).WriteMessage(websocket.TextMessage, []byte(backMsg))
	if err != nil {
		return err
	}
	return nil
}

func (h *Server) newConn(w http.ResponseWriter, r *http.Request) {
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		gstool.FmtPrintlnLog("Error during connection upgradation:", err)
		return
	}
	clientId := h.GetClientFunc(r)
	h.ClientConnMap.Set(clientId, conn)
	defer func(conn *websocket.Conn, clientId string) {
		err := conn.Close()
		if err != nil {
			gstool.FmtPrintlnLog(`建立连接失败%s`, err.Error())
		}
		h.ClientConnMap.Del(clientId)
	}(conn, clientId)
	//建立链接成功回调
	h.ConnectFunc(clientId, conn)
	// The event loop
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			gstool.FmtPrintlnLog("Error during message reading:%s", err)
			if h.CloseFunc != nil {
				h.CloseFunc(clientId)
			}
			break
		}
		if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
			h.ReceMsgFunc(clientId, cast.ToString(message))
		} else if messageType == websocket.CloseMessage {
			if h.CloseFunc != nil {
				h.CloseFunc(clientId)
			}
			break
		} else if messageType == websocket.PingMessage {
			err = conn.WriteMessage(websocket.PongMessage, []byte(`pong`))
			if err != nil {
				if h.CloseFunc != nil {
					h.CloseFunc(clientId)
				}
				break
			}
		}
	}
}
