package main

import (
	"dev_tool/internal/app/dtool/wailsapp"
	"embed"
	"flag"
	"io/fs"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

var ConfigFile string

const (
	// desktopWindowDefaultHeight 表示无法获取屏幕信息时的默认窗口高度。 // Defines the fallback window height when screen information is unavailable.
	desktopWindowDefaultHeight = 900
	// desktopWindowMinHeight 表示桌面窗口允许的最小高度。 // Defines the minimum allowed height for the desktop window.
	desktopWindowMinHeight = 700
	// desktopWindowFrameReserveHeight 为窗口标题栏和系统边框预留安全空间。 // Reserves safe space for the native title bar and window frame.
	desktopWindowFrameReserveHeight = 64
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	defaultConfigFile := os.Getenv("DTOOL_CONFIG_FILE")
	if defaultConfigFile == "" {
		defaultConfigFile = "config"
	}
	flag.StringVar(&ConfigFile, "ConfigFile", defaultConfigFile, "配置文件名 / Config file name")
	flag.Parse()

	// 显式切到 dist 子目录，避免 Wails 3 资源根目录解析错位。
	// Explicitly serve the dist subtree so Wails 3 resolves asset paths from the frontend bundle root.
	distAssets, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		panic(err)
	}

	// BundledAssetFileServer 在生产模式服务内嵌资源，在开发模式可自动接管 FRONTEND_DEVSERVER_URL。
	// BundledAssetFileServer serves embedded assets in production and automatically proxies FRONTEND_DEVSERVER_URL in development.
	desktopApp := wailsapp.NewDesktopApp(ConfigFile)
	app := application.New(application.Options{
		Name:        "dtool",
		Description: "dtool desktop client",
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(distAssets),
		},
		OnShutdown: desktopApp.Shutdown,
	})

	primaryScreen := app.Screen.GetPrimary()
	windowHeight, windowMinHeight := getDesktopWindowLayout(primaryScreen, desktopWindowMinHeight)

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "dtool",
		Width:            1400,
		Height:           windowHeight,
		MinWidth:         1100,
		MinHeight:        windowMinHeight,
		BackgroundColour: application.NewRGBA(255, 255, 255, 255),
		URL:              "/",
		InitialPosition:  application.WindowCentered,
	})
	desktopApp.BindRuntime(app, window)
	window.OnWindowEvent(events.Common.WindowRuntimeReady, func(_ *application.WindowEvent) {
		desktopApp.DomReady()
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}

// getDesktopWindowLayout 按主屏工作区高度生成桌面窗口高度与最小高度，目标为 90%，同时为系统标题栏预留空间。 // Calculates the desktop window height and minimum height from the primary screen work area, targeting 90% while reserving space for the native frame.
func getDesktopWindowLayout(primaryScreen *application.Screen, preferredMinHeight int) (height int, minHeight int) {
	if primaryScreen == nil || primaryScreen.WorkArea.Height <= 0 {
		return desktopWindowDefaultHeight, preferredMinHeight
	}
	maxVisibleHeight := primaryScreen.WorkArea.Height - desktopWindowFrameReserveHeight
	if maxVisibleHeight <= 0 {
		return desktopWindowDefaultHeight, preferredMinHeight
	}
	if preferredMinHeight > maxVisibleHeight {
		preferredMinHeight = maxVisibleHeight
	}
	targetHeight := primaryScreen.WorkArea.Height * 9 / 10
	// 先按 90% 取目标高度，再收敛到可见范围内，避免缩放场景下标题栏被顶出屏幕。 // Start from the 90% target, then clamp into the visible range so DPI scaling does not push the title bar off-screen.
	if targetHeight < preferredMinHeight {
		targetHeight = preferredMinHeight
	}
	if targetHeight > maxVisibleHeight {
		targetHeight = maxVisibleHeight
	}
	return targetHeight, preferredMinHeight
}
