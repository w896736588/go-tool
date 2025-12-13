package main

import (
	"dev_tool/internal/app/dtool"
	"flag"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

var ConfigFile string

func main() {
	flag.StringVar(&ConfigFile, `ConfigFile`, `config`, "是否是开发环境")
	flag.Parse()
	dtool.InitBase(ConfigFile)
	dtool.InitComponent()
	gstool.CpuSetUsePercent(0.6)
	gstool.SignalDefault()
	dtool.Stop()
}
