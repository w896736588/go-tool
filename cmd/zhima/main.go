package main

import (
	"gitee.com/Sxiaobai/gs/gstool"
	"xkf_tool/internal/app/zhima"
)

func main() {
	zhima.InitBase()
	gstool.CpuSetUsePercent(0.6)
	gstool.SignalDefault()
	zhima.Stop()
}
