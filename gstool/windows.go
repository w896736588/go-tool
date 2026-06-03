//go:build windows

package gstool

import (
	"errors"
	"github.com/spf13/cast"
	"golang.org/x/sys/windows"
	"syscall"
)

type WindowsScreen struct {
	ScreenWidth  int
	ScreenHeight int
	WorkWidth    int
	WorkHeight   int
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

var (
	user32                     = windows.NewLazySystemDLL("user32.dll")
	getDpiForWindowProc        = user32.NewProc("GetDpiForWindow")
	getDesktopWindowProc       = user32.NewProc("GetDesktopWindow")
	getSystemMetricsForDpiProc = user32.NewProc("GetSystemMetricsForDpi")
)

// WindowsWorkScreen 获取屏幕尺寸
func WindowsWorkScreen() *WindowsScreen {
	dpiRate, dpiErr := GetDpiForWindow()
	if dpiErr != nil {
		dpiRate = 1
	}
	FmtPrintlnLogTime(`dapiRate %#v`, dpiRate)
	// 定义常量
	SmCxscreen := 0      // 屏幕的宽度（以像素为单位）
	SmCyscreen := 1      // 屏幕的高度（以像素为单位）
	SmCxfullscreen := 16 // 工作区域的宽度（以像素为单位）
	SmCyfullscreen := 17 // 工作区域的高度（以像素为单位）
	// 调用 GetSystemMetrics 获取屏幕尺寸
	var user32 = syscall.NewLazyDLL("user32.dll")
	var procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	screenWidth, _, _ := procGetSystemMetrics.Call(uintptr(SmCxscreen))
	screenHeight, _, _ := procGetSystemMetrics.Call(uintptr(SmCyscreen))
	workAreaWidth, _, _ := procGetSystemMetrics.Call(uintptr(SmCxfullscreen))
	workAreaHeight, _, _ := procGetSystemMetrics.Call(uintptr(SmCyfullscreen))
	return &WindowsScreen{
		ScreenWidth:  cast.ToInt(float32(uint32(screenWidth)) * dpiRate),
		ScreenHeight: int(screenHeight),
		WorkWidth:    cast.ToInt(float32(uint32(workAreaWidth)) * dpiRate),
		WorkHeight:   int(workAreaHeight),
	}
}

// GetDpiForWindow 获取屏幕缩放率
func GetDpiForWindow() (float32, error) {
	hwnd, _, _ := getDesktopWindowProc.Call()
	if hwnd == 0 {
		return 0, errors.New(`获取窗口失败`)
	}
	ret, _, _ := getDpiForWindowProc.Call(hwnd)
	if ret == 0 {
		return 0, Error(`获取缩放率失败`)
	}
	dpiUnit32 := uint32(ret)
	dpiFloat := cast.ToFloat32(dpiUnit32)
	return cast.ToFloat32(dpiFloat / 96), nil
}
