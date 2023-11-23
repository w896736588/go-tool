package base_module

import (
	"gitee.com/Sxiaobai/gs/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func (h *Global) GinInit(host, port string) *gsgin.GSGin {
	if h.gin != nil {
		return h.gin
	}
	gsGin := &gsgin.GSGin{
		Host: host,
		Port: cast.ToInt(port),
	}
	gsGin.CreateRouter()
	h.gin = gsGin
	return h.gin
}

func (h *Global) GinSetAllowCrossDomain() {
	h.gin.SetAllow()
}

func (h *Global) GinPost(route string, call ...gin.HandlerFunc) {
	h.gin.GinH.POST(route, call...)
}

func (h *Global) GinGet(route string, call ...gin.HandlerFunc) {
	h.gin.GinH.GET(route, call...)
}

func (h *Global) GinStatic(route, root string) {
	h.gin.GinH.Static(route, root)
}

func (h *Global) GinStaticFile(relativePath, filepath string) {
	h.gin.GinH.StaticFile(relativePath, filepath)
}

func (h *Global) GinLoadHTMLFiles(file ...string) {
	h.gin.GinH.LoadHTMLFiles(file...)
}

func (h *Global) GinRun() {
	h.gin.Run()
}

func (h *Global) GinStop(waitSecond int64) error {
	return h.gin.Stop(waitSecond)
}
