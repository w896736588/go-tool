package gsssh

import (
	"io"
	"net"
	"time"

	"github.com/w896736588/go-tool/gstool"
)

type SshBridge struct {
	sshHandle      *Ssh
	localListener  net.Listener
	remoteConn     net.Conn
	localConn      net.Conn
	targetHostPort string
	localHostPort  string
}

func NewSshBridge(sshHandle *Ssh) *SshBridge {
	return &SshBridge{
		sshHandle: sshHandle,
	}
}

// RunBridge execute and forward to the local port
// Local listening port, all connections to this port will be forwarded to the destination address
func (h *SshBridge) RunBridge(targetHostPort string) (string, error) {
	if h.sshHandle.client == nil {
		connectionErr := h.sshHandle.ConnectAuthPassword()
		if connectionErr != nil {
			return ``, connectionErr
		}
	}
	h.targetHostPort = targetHostPort
	//监听本地端口
	listenLocalErr := h.startListenLocal()
	if listenLocalErr != nil {
		return ``, listenLocalErr
	}
	//监听目标端口
	listenRemoteErr := h.createRemoteConn()
	if listenRemoteErr != nil {
		return ``, listenRemoteErr
	}
	//创建链接
	go h.createConn()
	return h.localHostPort, nil
}

func (h *SshBridge) startListenLocal() error {
	var localListenerErr error
	h.localListener, localListenerErr = net.Listen("tcp", "127.0.0.1:0")
	if localListenerErr != nil {
		return gstool.Error("监听本地端口失败: %s", localListenerErr.Error())
	}
	h.localHostPort = h.localListener.Addr().String()
	return nil
}

func (h *SshBridge) createRemoteConn() error {
	var remoteConnErr error
	h.remoteConn, remoteConnErr = h.sshHandle.client.Dial("tcp", h.targetHostPort)
	if remoteConnErr != nil {
		return gstool.Error("连接目标端口失败: %s", remoteConnErr.Error())
	}
	return nil
}

func (h *SshBridge) createLocalConn() error {
	var localConnErr error
	h.localConn, localConnErr = h.localListener.Accept()
	if localConnErr != nil {
		gstool.FmtPrintlnLogTime(`接收链接到本地监听端口(%s)失败 %s`, h.localHostPort, localConnErr.Error())
		return localConnErr
	}
	return nil
}

// bridgeListenLocal 等待本地端口连接
// 注意：只有真正发送请求的时候（例如ping） 才会开始执行 如果仅仅是监听本地端口是不会执行后续的东西的
func (h *SshBridge) createConn() {
	localErr := h.createLocalConn()
	if localErr != nil {
		gstool.FmtPrintlnLogTime(`创建local conn失败 %s`, localErr.Error())
		return
	}
	// 复制流量
	h.transferCopy(true, true)
}

// The transfer function runs in the goroutine and is used to copy data between two connections
func (h *SshBridge) transferCopy(boolLTR, boolRTL bool) {
	go func() {
		h.ioCopyLocalToRemote()
	}()
	go func() {
		if boolRTL {
			h.ioCopyRemoteToLocal()
		}
	}()
}

func (h *SshBridge) ioCopyLocalToRemote() {
	defer func() {
		if r := recover(); r != nil {
			gstool.FmtPrintlnLogTime("ioCopyLocalToRemote warn:%v", r)
		}
	}()
	_, _ = io.Copy(h.localConn, h.remoteConn)
}

func (h *SshBridge) ioCopyRemoteToLocal() {
	defer func() {
		if r := recover(); r != nil {
			gstool.FmtPrintlnLogTime("ioCopyRemoteToLocal warn:%v", r)
		}
	}()
	ret, _ := io.Copy(h.remoteConn, h.localConn)
	if ret > 0 {
		h.CloseBridge()
	}
}

func (h *SshBridge) CloseBridge() {
	_ = h.remoteConn.Close()
	_ = h.localConn.Close()
	time.Sleep(time.Second * 2)
	localErr := h.createLocalConn()
	if localErr != nil {
		gstool.FmtPrintlnLogTime(`local err %s`, localErr.Error())
	}
	err := h.createRemoteConn()
	if err != nil {
		gstool.FmtPrintlnLogTime(`重连失败 %s`, err.Error())
	}
	h.transferCopy(false, true)
}
