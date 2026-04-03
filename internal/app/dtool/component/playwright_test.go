package component

import (
	"dev_tool/internal/app/dtool/define"
	"testing"
)

func TestNewTPlaywrightUsesEnvConfig(t *testing.T) {
	originalEnv := EnvClient
	t.Cleanup(func() {
		EnvClient = originalEnv
	})

	EnvClient = &define.Env{
		LogPath:            t.TempDir(),
		WebkitDownloadPath: `C:\tmp\webkit_download`,
		WebkitDriverPath:   `C:\tmp\webkit_driver`,
		NodePath:           `node`,
	}

	client := NewTPlaywright()
	if client == nil {
		t.Fatal("NewTPlaywright() returned nil")
	}
	t.Cleanup(func() {
		_ = client.Log.Close()
	})
	if client.DownloadPath != EnvClient.WebkitDownloadPath {
		t.Fatalf("DownloadPath = %q, want %q", client.DownloadPath, EnvClient.WebkitDownloadPath)
	}
	if client.Log == nil {
		t.Fatal("NewTPlaywright() returned client without logger")
	}
}
