package base_module

import (
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gstool"
)

type Component struct {
	ShellClient       *gstool.GsShellPush
	RedisClient       *gsdb.GsRedis
	XkfMysqlClient    *gsdb.GsMysql
	AppUrlMysqlClient *gsdb.GsMysql
	Global            *Global
	ReqMap            map[string]any
	Logger            *gstool.GsSlog
	Encrypt           *gstool.Encrypt
}
