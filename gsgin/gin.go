package gsgin

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

type GSGin struct {
	Host   string
	Port   int
	GinH   *gin.Engine
	GinSrv *http.Server
}

// CreateDefaultRouter creates 默认组件中会开启日志
func (h *GSGin) CreateDefaultRouter() {
	h.GinH = gin.Default()
}

// CreateRouter 创建自定义路由 默认增加错误捕获
func (h *GSGin) CreateRouter() {
	h.GinH = gin.New()
	h.GinH.Use(gin.Recovery()) //错误捕获 错误将写入错误输出
}

// SetAllow 允许跨域访问
func (h *GSGin) SetAllow() {
	h.GinH.Use(h.setAllow())
}

// Run 运行
func (h *GSGin) Run() {
	h.GinSrv = &http.Server{
		Addr:    fmt.Sprintf(`%s:%d`, h.Host, h.Port),
		Handler: h.GinH,
	}
	gstool.FmtPrintlnLog(`gin Addr %s`, fmt.Sprintf(`%s:%d`, h.Host, h.Port))
	go func() {
		if err := h.GinSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err.Error())
		}
	}()
}

// Stop 平滑停止
func (h *GSGin) Stop(waitSecond int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(waitSecond)*time.Second)
	defer cancel()
	if err := h.GinSrv.Shutdown(ctx); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return nil
	}
}

// SetAllow 允许跨域访问具体处理
func (h *GSGin) setAllow() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
		}
		// 处理请求
		c.Next()
	}
}

func (h *GSGin) SseRoute(route string, allowOrigin bool,
	openFunc func(urlValues url.Values, stopC chan int, c *gin.Context) (*Sse, error), closeFunc func(sse *Sse)) {
	h.GinH.GET(route, func(c *gin.Context) {
		method := c.Request.Method
		if allowOrigin {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "*")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
			if method == "OPTIONS" {
				c.AbortWithStatus(http.StatusOK)
				return
			}
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")

		stopC := make(chan int, 1)
		clientGone := c.Writer.CloseNotify()
		sse, err := openFunc(GinGetParams(c), stopC, c)
		if err != nil {
			c.String(200, err.Error())
			return
		}
		defer func() {
			closeFunc(sse)
		}()
		for {
			select {
			case <-stopC:
				return
			case <-clientGone:
				close(stopC)
				return
			}
		}
	})
}
