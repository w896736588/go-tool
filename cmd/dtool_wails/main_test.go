package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestStartupSplashPageLoadsWailsRuntime(t *testing.T) {
	indexPath := filepath.Join(`frontend`, `dist`, `index.html`)
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", indexPath, err)
	}

	if !strings.Contains(string(content), `/wails/runtime.js`) {
		t.Fatalf("startup splash page should load /wails/runtime.js so WindowRuntimeReady can fire")
	}
}

func TestGetDesktopWindowLayoutUsesNinetyPercentOfPrimaryWorkArea(t *testing.T) {
	screen := &application.Screen{
		WorkArea: application.Rect{
			Height: 1000,
		},
	}

	gotHeight, gotMinHeight := getDesktopWindowLayout(screen, desktopWindowMinHeight)
	if gotHeight != 900 || gotMinHeight != desktopWindowMinHeight {
		t.Fatalf("getDesktopWindowLayout() = (%d, %d), want (%d, %d)", gotHeight, gotMinHeight, 900, desktopWindowMinHeight)
	}
}

func TestGetDesktopWindowLayoutClampsHeightIntoVisibleArea(t *testing.T) {
	screen := &application.Screen{
		WorkArea: application.Rect{
			Height: 720,
		},
	}

	gotHeight, gotMinHeight := getDesktopWindowLayout(screen, desktopWindowMinHeight)
	wantVisibleHeight := 720 - desktopWindowFrameReserveHeight
	if gotHeight != wantVisibleHeight || gotMinHeight != wantVisibleHeight {
		t.Fatalf("getDesktopWindowLayout() = (%d, %d), want (%d, %d)", gotHeight, gotMinHeight, wantVisibleHeight, wantVisibleHeight)
	}
}
