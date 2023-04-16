package xkf_tool_gin

import (
	"gitee.com/Sxiaobai/gs/gsgin"
	"github.com/gin-gonic/gin"
	"net/http"
)

func InitRouter(host, port string) error {
	gsGin := gsgin.GSGin{
		Host: host,
		Port: port,
	}
	gsGin.CreateRouter()
	gsGin.SetAllow()

	//REDIS接口
	//展示所有的
	gsGin.GinH.POST(`/api/redis/list`, RedisList)
	//查询某个key
	gsGin.GinH.POST(`/api/search`, Search)
	//模糊搜索key
	gsGin.GinH.POST(`/api/keys`, Keys)
	//批量获取缓存类型
	gsGin.GinH.POST(`/api/keys/type`, KeysType)
	//获取key类型
	gsGin.GinH.POST(`/api/key/type`, GetKeyType)
	//序列化和反序列化
	gsGin.GinH.POST(`/api/serialize`, PhpSerialize)
	gsGin.GinH.POST(`/api/unserialize`, PhpUnSerialize)
	//保存string
	gsGin.GinH.POST(`/api/save/string`, SaveString)
	//删除key
	gsGin.GinH.POST(`/api/del/key`, DelKey)
	//删除sub key
	gsGin.GinH.POST(`/api/del/sub`, DelSub)
	//更改ttl
	gsGin.GinH.POST(`/api/edit/ttl`, EditTtl)
	//删除所有缓存
	gsGin.GinH.POST(`/api/delete/all`, DelAllKey)
	//创建缓存
	gsGin.GinH.POST(`/api/create`, CreateCache)
	//编辑二级缓存
	gsGin.GinH.POST(`/api/edit/sub`, EditSub)
	//找到所有启用的消费者服务
	//router.POST(`/api/supervisor/status`, SupervisorStatus)

	//shell exec
	gsGin.GinH.POST(`/api/shell/exec`, ShellExec)

	//前端页面
	gsGin.GinH.Static(`/static`, `./views/dist/static`)
	gsGin.GinH.LoadHTMLFiles(`views/dist/index.html`)
	gsGin.GinH.GET("/", func(context *gin.Context) {
		context.HTML(200, "index.html", nil)
	})
	return gsGin.GinH.Run()
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header(`Access-Control-Allow-Private-Network`, `*`)
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
		}
		// 处理请求
		c.Next()
	}
}
