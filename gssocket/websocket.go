package gssocket

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

type Server struct {
	Host          string
	Uri           string
	ReceMsgFunc   func(string, string)
	ConnectFunc   func(string, *websocket.Conn)
	CloseFunc     func(string)
	GetClientFunc func(r *http.Request) string
	AllowOrigin   bool

	ReadTimeout  time.Duration // 读超时，默认 120s
	WriteTimeout time.Duration // 写超时，默认 15s

	ClientConnMap *gstool.HighMap
	writeLockMap  *gstool.HighMap // 每个连接的写锁
	clientIds     []string        // 已连接的 clientId 列表，用于 Shutdown 遍历关闭
	clientIdLock  sync.Mutex      // clientIds 的锁

	upgrader websocket.Upgrader
	server   *http.Server
}

func (h *Server) Start() {
	if h.ReadTimeout <= 0 {
		h.ReadTimeout = 120 * time.Second
	}
	if h.WriteTimeout <= 0 {
		h.WriteTimeout = 15 * time.Second
	}

	h.ClientConnMap = gstool.HighMapCreate(16)
	h.writeLockMap = gstool.HighMapCreate(16)
	h.upgrader = websocket.Upgrader{}
	if h.AllowOrigin {
		h.upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc(h.Uri, h.newConn)
	h.server = &http.Server{Addr: h.Host, Handler: mux}

	err := h.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		gstool.FmtPrintlnLog(`websocket 服务启动失败: %s`, err.Error())
	}
}

// Shutdown 优雅关闭：先向所有 WebSocket 连接发送 Close 帧，再关闭 HTTP 服务
func (h *Server) Shutdown(ctx context.Context) error {
	h.clientIdLock.Lock()
	ids := make([]string, len(h.clientIds))
	copy(ids, h.clientIds)
	h.clientIdLock.Unlock()

	for _, id := range ids {
		conn, ok := h.ClientConnMap.Get(id)
		if !ok {
			continue
		}
		mu, _ := h.writeLockMap.Get(id)
		if mu != nil {
			mu.(*sync.Mutex).Lock()
			conn.(*websocket.Conn).WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			mu.(*sync.Mutex).Unlock()
		}
	}

	if h.server == nil {
		return nil
	}
	return h.server.Shutdown(ctx)
}

func (h *Server) WriteMessage(clientId, backMsg string) error {
	conn, ok := h.ClientConnMap.Get(clientId)
	if !ok {
		return errors.New("不存在的链接")
	}

	mu, _ := h.writeLockMap.Get(clientId)
	if mu == nil {
		return errors.New("写锁不存在")
	}
	mu.(*sync.Mutex).Lock()
	defer mu.(*sync.Mutex).Unlock()

	conn.(*websocket.Conn).SetWriteDeadline(time.Now().Add(h.WriteTimeout))
	return conn.(*websocket.Conn).WriteMessage(websocket.TextMessage, []byte(backMsg))
}

func (h *Server) newConn(w http.ResponseWriter, r *http.Request) {
	if h.GetClientFunc == nil {
		gstool.FmtPrintlnLog("GetClientFunc 未设置，拒绝连接")
		return
	}
	clientId := h.GetClientFunc(r)
	if clientId == "" {
		gstool.FmtPrintlnLog("clientId 为空，拒绝连接")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		gstool.FmtPrintlnLog("WebSocket 升级失败: %s", err.Error())
		return
	}

	h.ClientConnMap.Set(clientId, conn)
	h.writeLockMap.Set(clientId, &sync.Mutex{})
	h.clientIdLock.Lock()
	h.clientIds = append(h.clientIds, clientId)
	h.clientIdLock.Unlock()

	defer func() {
		h.clientIdLock.Lock()
		for i, id := range h.clientIds {
			if id == clientId {
				h.clientIds = append(h.clientIds[:i], h.clientIds[i+1:]...)
				break
			}
		}
		h.clientIdLock.Unlock()
		h.writeLockMap.Del(clientId)
		h.ClientConnMap.Del(clientId)
		if err := conn.Close(); err != nil {
			gstool.FmtPrintlnLog("关闭连接失败 clientId=%s: %s", clientId, err.Error())
		}
	}()

	if h.ConnectFunc != nil {
		h.ConnectFunc(clientId, conn)
	}

	// 设置读超时和 Pong 处理器实现心跳
	conn.SetReadDeadline(time.Now().Add(h.ReadTimeout))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(h.ReadTimeout))
		return nil
	})

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			gstool.FmtPrintlnLog("读取消息失败 clientId=%s: %s", clientId, err.Error())
			if h.CloseFunc != nil {
				h.CloseFunc(clientId)
			}
			break
		}
		conn.SetReadDeadline(time.Now().Add(h.ReadTimeout))

		switch messageType {
		case websocket.TextMessage, websocket.BinaryMessage:
			if h.ReceMsgFunc != nil {
				h.ReceMsgFunc(clientId, cast.ToString(message))
			}
		case websocket.PingMessage:
			mu, _ := h.writeLockMap.Get(clientId)
			if mu != nil {
				mu.(*sync.Mutex).Lock()
				err = conn.WriteMessage(websocket.PongMessage, nil)
				mu.(*sync.Mutex).Unlock()
			}
			if err != nil {
				if h.CloseFunc != nil {
					h.CloseFunc(clientId)
				}
				return
			}
		}
	}
}
