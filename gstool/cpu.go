package gstool

import (
	"math"
	"runtime"
)

// CpuGetNum 拿到cpu核数
func CpuGetNum() int {
	return runtime.NumCPU()
}

// CpuSetUseNum 设置使用的cpu核数
func CpuSetUseNum(runNum int) {
	runtime.GOMAXPROCS(runNum)
}

// CpuSetUsePercent 设置最高使用几个核
func CpuSetUsePercent(percent float64) {
	cpuNum := runtime.NumCPU()
	setCpu := int(math.Ceil(float64(cpuNum) * percent))
	CpuSetUseNum(setCpu)
}
