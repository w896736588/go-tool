package gsws

// SetReceiveFunc 设置回调方法
func (h *WsConn) SetReceiveFunc(receiveFunc func(message string)) {
	h.receiveMsgFunc = receiveFunc
}

// SetConnCloseFunc 连接关闭回调
func (h *Server) SetConnCloseFunc(closeFunc func(clientId string)) {
	h.connCloseFunc = closeFunc
}

// SetGClientIdFunc 获取clientId回调
func (h *Server) SetGClientIdFunc(getClientIdFunc func() string) {
	h.getClientIdFunc = getClientIdFunc
}

// SetNewConnFunc 新连接回调
func (h *Server) SetNewConnFunc(newCliConnFunc func(wsConn *WsConn, queryParams map[string]any)) {
	h.newConnFunc = newCliConnFunc
}
