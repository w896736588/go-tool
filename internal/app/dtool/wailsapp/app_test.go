package wailsapp

import (
	"os"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type fakeRedirectWindow struct {
	lastURL string
}

func (h *fakeRedirectWindow) SetURL(url string) application.Window {
	h.lastURL = url
	return nil
}

func TestOpenBackendURLRedirectsWindowToBackend(t *testing.T) {
	window := &fakeRedirectWindow{}
	app := &DesktopApp{
		window: window,
	}

	app.openBackendURL(`http://localhost:17170/`)

	if window.lastURL != `http://localhost:17170/` {
		t.Fatalf("lastURL = %q, want %q", window.lastURL, `http://localhost:17170/`)
	}
}

func TestIsExternalBackendManagedReadsEnvFlag(t *testing.T) {
	oldValue, hadOldValue := os.LookupEnv(`DTOOL_WAILS_DEV_EXTERNAL_BACKEND`)
	t.Cleanup(func() {
		if hadOldValue {
			_ = os.Setenv(`DTOOL_WAILS_DEV_EXTERNAL_BACKEND`, oldValue)
			return
		}
		_ = os.Unsetenv(`DTOOL_WAILS_DEV_EXTERNAL_BACKEND`)
	})

	_ = os.Setenv(`DTOOL_WAILS_DEV_EXTERNAL_BACKEND`, `1`)
	if !isExternalBackendManaged() {
		t.Fatalf(`isExternalBackendManaged() = false, want true`)
	}

	_ = os.Setenv(`DTOOL_WAILS_DEV_EXTERNAL_BACKEND`, `0`)
	if isExternalBackendManaged() {
		t.Fatalf(`isExternalBackendManaged() = true, want false`)
	}
}
