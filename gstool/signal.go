package gstool

import (
	"os"
	"os/signal"
	"syscall"
)

// SignalDefault 监听程序结束
func SignalDefault() os.Signal {
	sc := make(chan os.Signal)
	sl := []os.Signal{
		syscall.SIGHUP,  //热升级
		syscall.SIGINT,  //Ctrl+C
		syscall.SIGTERM, //结束程序
	}
	signal.Notify(sc, sl...)
	return <-sc
}
