package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"redis_manager/api/gin"
	"redis_manager/internal/app/xkf_tool"
	"syscall"
)

func main() {
	xkf_tool.InitConfig()
	router := gin.InitRouter()
	err := router.Run(fmt.Sprintf(`:%s`, xkf_tool.ConfigViper.GetString(`run.port`)))
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
