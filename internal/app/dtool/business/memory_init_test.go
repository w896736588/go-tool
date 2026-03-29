package business

import (
	"dev_tool/internal/app/dtool/component"
	"testing"

	"github.com/spf13/viper"
)

func TestReadMemoryConfigFromINI(t *testing.T) {
	oldViper := component.ConfigViper
	t.Cleanup(func() {
		component.ConfigViper = oldViper
	})

	v := newTestMemoryConfigViper()
	v.Set(`base.memoryDbPath`, `D:\repo\memory`)
	v.Set(`base.memoryDbFileName`, `memory.db`)
	v.Set(`base.memoryDbIsGitRepo`, true)
	component.ConfigViper = v

	got := ReadMemoryConfigFromINI()
	if got.Dir != `D:\repo\memory` {
		t.Fatalf("Dir = %q, want %q", got.Dir, `D:\repo\memory`)
	}
	if got.DBName != `memory.db` {
		t.Fatalf("DBName = %q, want %q", got.DBName, `memory.db`)
	}
	if got.DBPath != `D:\repo\memory\memory.db` {
		t.Fatalf("DBPath = %q, want %q", got.DBPath, `D:\repo\memory\memory.db`)
	}
	if !got.GitRepoEnabled {
		t.Fatalf("GitRepoEnabled = %v, want true", got.GitRepoEnabled)
	}
}

func TestReadMemoryConfigFromINIHandlesMissingConfig(t *testing.T) {
	oldViper := component.ConfigViper
	t.Cleanup(func() {
		component.ConfigViper = oldViper
	})

	component.ConfigViper = newTestMemoryConfigViper()
	got := ReadMemoryConfigFromINI()
	if got.Dir != `` || got.DBName != `` || got.DBPath != `` || got.GitRepoEnabled {
		t.Fatalf("expected empty memory config, got %+v", got)
	}
}

func newTestMemoryConfigViper() *viper.Viper {
	return viper.New()
}
