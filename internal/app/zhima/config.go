package zhima

import (
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"xkf_tool/base_module"
	"xkf_tool/internal/app/zhima/controller"
)

var AppName = `zhima`
var Logger *gstool.GsLogger
var RootPath string
var ConfigViper *viper.Viper
var Encrypt *gstool.Encrypt //全局加密
var Global *base_module.Global
var GlobalGin *base_module.Global //专为gin处理的

func InitBase() {
	var err error
	RootPath, err = gstool.GetRootPath()
	if err != nil {
		panic(err.Error())
	}
	getConfigViper()
	getEncrypt()
	getLogger()
	getGin()
	getGlobal()
}

func Stop() {
	err := GlobalGin.GinStop(10)
	if err != nil {
		Global.Error(err.Error())
	}
}
func getLogger() {
	Logger = gstool.CreateLogger(RootPath+`/logs/`+AppName, ``)
}
func getGlobal() {
	Global = &base_module.Global{}
	Global.SetLogger(Logger)
}
func getConfigViper() {
	ConfigViper = viper.New()
	ConfigViper.AddConfigPath(RootPath + `/config/` + AppName)
	ConfigViper.SetConfigName(`config`)
	ConfigViper.SetConfigType(`ini`)
	if err := ConfigViper.ReadInConfig(); err != nil {
		panic(err.Error())
	}
}
func getEncrypt() {
	Encrypt = &gstool.Encrypt{
		Key: ConfigViper.GetString(`encrypt.key`),
		Iv:  ConfigViper.GetString(`encrypt.iv`),
	}
}
func getGin() {
	host := ConfigViper.GetString(`run.host`)
	port := ConfigViper.GetString(`run.port`)
	GlobalGin := &base_module.Global{}
	GlobalGin.SetLogger(Logger)
	GlobalGin.GinInit(host, port)
	GlobalGin.GinSetAllowCrossDomain()
	GlobalGin.GinStatic(`/static`, `./views/dist/static`)
	GlobalGin.GinLoadHTMLFiles(`views/dist/index.html`)
	GlobalGin.GinGet(`/`, func(context *gin.Context) {
		context.HTML(200, `index.html`, nil)
	})
	GlobalGin.GinPost(`/api/Login`, controller.Login)                           //登录
	GlobalGin.GinPost(`/api/RegisterService`, controller.RegisterService)       //注册各类服务
	GlobalGin.GinPost(`/api/RedisAvailableList`, controller.RedisAvailableList) //可用的redis列表
	GlobalGin.GinPost(`/api/RedisSearch`, controller.RedisSearch)               //查询某个key
	GlobalGin.GinPost(`/api/RedisKeys`, controller.RedisKeys)                   //模糊搜索key
	GlobalGin.GinPost(`/api/RedisKeyType`, controller.RedisKeyType)             //批量获取缓存类型
	GlobalGin.GinPost(`/api/key/type`, GetKeyType)                              //获取key类型
	GlobalGin.GinPost(`/api/serialize`, PhpSerialize)                           //序列化和反序列化
	GlobalGin.GinPost(`/api/unserialize`, PhpUnSerialize)
	GlobalGin.GinPost(`/api/save/string`, SaveString) //保存string
	GlobalGin.GinPost(`/api/del/key`, DelKey)         //删除key
	GlobalGin.GinPost(`/api/del/sub`, DelSub)         //删除sub key
	GlobalGin.GinPost(`/api/edit/ttl`, EditTtl)       //更改ttl
	GlobalGin.GinPost(`/api/delete/all`, DelAllKey)   //删除所有缓存
	GlobalGin.GinPost(`/api/create`, CreateCache)     //创建缓存
	GlobalGin.GinPost(`/api/edit/sub`, EditSub)       //编辑二级缓存
	GlobalGin.GinPost(`/api/shell/exec`, ShellExec)
	GlobalGin.GinRun()
}
