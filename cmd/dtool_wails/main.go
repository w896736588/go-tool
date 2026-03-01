package main

import (
	"dev_tool/internal/app/dtool/wailsapp"
	"embed"
	"flag"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

var ConfigFile string

//go:embed frontend/dist
var assets embed.FS

func main() {
	defaultConfigFile := os.Getenv("DTOOL_CONFIG_FILE")
	if defaultConfigFile == "" {
		defaultConfigFile = "config"
	}
	flag.StringVar(&ConfigFile, `ConfigFile`, defaultConfigFile, "配置文件名")
	flag.Parse()

	app := wailsapp.NewDesktopApp(ConfigFile)
	err := wails.Run(&options.App{
		Title:            "dtool",
		Width:            1400,
		Height:           900,
		MinWidth:         1100,
		MinHeight:        700,
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.Startup,
		OnDomReady:       app.DomReady,
		OnShutdown:       app.Shutdown,
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
	})
	if err != nil {
		panic(err)
	}
}
