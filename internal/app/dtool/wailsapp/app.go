package wailsapp

import (
	"context"
	"dev_tool/internal/app/dtool"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// DesktopApp 管理桌面端生命周期与后端服务启动。
type DesktopApp struct {
	ctx        context.Context
	configFile string
	bootOnce   sync.Once
}

// NewDesktopApp 创建桌面应用实例。
func NewDesktopApp(configFile string) *DesktopApp {
	return &DesktopApp{configFile: configFile}
}

// Startup 在应用启动时保存上下文。
func (a *DesktopApp) Startup(ctx context.Context) {
	a.ctx = ctx
}

// DomReady 在前端可用后异步拉起后端并跳转到本地页面。
func (a *DesktopApp) DomReady(ctx context.Context) {
	go a.bootBackendAndOpen()
}

// Shutdown 在桌面应用退出时关闭后端资源。
func (a *DesktopApp) Shutdown(ctx context.Context) {
	dtool.Stop()
}

// bootBackendAndOpen 保证后端只初始化一次，并在端口就绪后跳转。
func (a *DesktopApp) bootBackendAndOpen() {
	a.bootOnce.Do(func() {
		dtool.InitBase(a.configFile)
		go dtool.InitComponent()

		if !waitPortReady(dtool.GetPrimaryPort(), 30*time.Second) {
			runtime.LogErrorf(a.ctx, "后端端口启动超时: %s", dtool.GetPrimaryPort())
			return
		}

		targetURL := dtool.GetPrimaryURL()
		runtime.WindowExecJS(a.ctx, fmt.Sprintf("window.location.replace(%q)", targetURL))
	})
}

// waitPortReady 轮询检测端口是否可建立连接。
func waitPortReady(port string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	addr := "127.0.0.1:" + port
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 600*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false
}
