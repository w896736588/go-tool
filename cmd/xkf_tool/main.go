package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"xkf_tool/internal/app/xkf_tool"
	"xkf_tool/internal/app/xkf_tool_gin"
)

func main() {
	xkf_tool.InitConfig()
	//host port
	port := xkf_tool.ConfigViper.GetString(`run.port`)
	host := `localhost`
	err := xkf_tool_gin.InitRouter(host, port)
	if err != nil {
		log.Errorf(`%s`, err.Error())
		return
	}
	sc := make(chan os.Signal)
	sl := []os.Signal{
		syscall.SIGHUP,  //热升级
		syscall.SIGINT,  //Ctrl+C
		syscall.SIGTERM, //结束程序
	}
	signal.Notify(sc, sl...)
}
