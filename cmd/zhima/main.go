package main

import (
	"dev_tool/internal/app/zhima"
	"gitee.com/Sxiaobai/gs/gstool"
)

func main() {
	zhima.InitBase()
	gstool.CpuSetUsePercent(0.6)
	gstool.SignalDefault()
}
