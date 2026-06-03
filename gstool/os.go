package gstool

import (
	"os/exec"
	"runtime"
)

type GsOsType string

const GsWindows GsOsType = `windows`
const GsLinux GsOsType = `linux`
const GsMac GsOsType = `darwin`

type GsOs struct {
	System GsOsType
}

func NewGsOs() *GsOs {
	return &GsOs{
		System: GsOsType(runtime.GOOS), //windows linux darwin ....
	}
}

func (h *GsOs) OpenFileWindows(title, filePath string) error {
	cmd := exec.Command("cmd", "/C", "start", title, filePath)
	return cmd.Start()
}

func (h *GsOs) OpenDirWindows(dirPath string) error {
	cmd := exec.Command("explorer", dirPath)
	return cmd.Start()
}
