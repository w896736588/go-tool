package gsws

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/w896736588/go-tool/gstool"
)

type Config struct {
	Host string
	Port string
	Uri  string //路由

	ReadBufferSize  int  //IO读缓冲区
	WriteBufferSize int  //IO写缓冲区
	CheckOrigin     bool //是否检查跨域 true：不允许跨域，false：允许跨域

	MaxLiveTime      int64         //连接最大生存时间 心跳要小于这个值
	MaxWriteWaitTime time.Duration //推送消息最大等待时间
	MaxReadWaitTime  time.Duration //读取消息最大的等待时间
	MaxMessageSize   int64         //消息最大byte

	MaxClient int //最大的客户端数

}

// Server 服务总配置
type Server struct {
	getClientIdFunc func() string                                  //获取client Id回调
	connCloseFunc   func(string)                                   //连接关闭回调
	newConnFunc     func(conn *WsConn, queryParams map[string]any) //新连接回调
	ginH            *gin.Engine
	upGraderH       *websocket.Upgrader
	config          Config //配置

	currentClient int                //当前的连接数
	startTime     int64              //启动时间 用于生成连接ID
	cliConnMap    map[string]*WsConn //所有客户端的连接
	syncLock      sync.RWMutex
}

// Start 启动
func (h *Server) Start() error {
	h.ginH = gin.Default()
	h.ginH.GET(h.config.Uri, h.connect)
	h.upGraderH = &websocket.Upgrader{
		ReadBufferSize:  h.config.ReadBufferSize,
		WriteBufferSize: h.config.WriteBufferSize,
	}
	h.startTime = time.Now().Unix()
	h.cliConnMap = make(map[string]*WsConn)
	//是否允许跨域
	if !h.config.CheckOrigin {
		h.upGraderH.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}
	return h.ginH.Run(h.config.Host + `:` + h.config.Port)
}

// SetConfig 设置配置
func (h *Server) SetConfig(conf Config) {
	h.config = conf
}

// Connect 连接 升级为websocket
func (h *Server) connect(c *gin.Context) {
	queryParams := h.getQueryParams(c)
	//升级为websocket
	conn, err := h.upGraderH.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		gstool.FmtPrintlnLog(`建立连接错误 %s`, err.Error())
		return
	}
	if h.currentClient > h.config.MaxClient {
		gstool.FmtPrintlnLog(`当前连接数已达上限`)
		return
	}
	//当前连接数+1
	h.currentClient++
	//构造客户端连接
	conn.SetReadLimit(h.config.MaxMessageSize) //设置消息最大字节

	cliConn := WsConn{
		conn:          conn,
		readChan:      make(chan string, 1000),
		writeChan:     make(chan string, 1000),
		mutex:         sync.Mutex{},
		isClosed:      false,
		closeChan:     make(chan byte),
		clientId:      h.getClientIdFunc(),
		serverHandle:  h,
		lastHeartTime: time.Now().Unix(),
	}
	h.addCliConn(&cliConn)
	//新连接回调
	h.newConnFunc(&cliConn, queryParams)
	// 处理器,发送定时信息，避免意外关闭
	go cliConn.doBusinessLoop()
	// 读协程
	go cliConn.wsReadLoop()
	// 写协程
	go cliConn.wsWriteLoop()
}

// 获取请求的参数
func (h *Server) getQueryParams(c *gin.Context) map[string]any {
	query := c.Request.URL.Query()
	var queryMap = make(map[string]any, len(query))
	for k := range query {
		queryMap[k] = c.Query(k)
	}
	return queryMap
}

// 增加一个连接
func (h *Server) addCliConn(cliConn *WsConn) {
	h.syncLock.Lock()
	h.cliConnMap[cliConn.clientId] = cliConn
	h.syncLock.Unlock()
}

// 关闭连接
func (h *Server) removeCliConn(cliConn *WsConn) {
	h.syncLock.Lock()
	delete(h.cliConnMap, cliConn.clientId)
	h.currentClient--
	h.connCloseFunc(cliConn.clientId)
	h.syncLock.Unlock()

}
