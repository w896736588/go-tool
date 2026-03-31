package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"errors"
	"os"
	"path/filepath"
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}
	if got.Dir != filepath.Join(homeDir, `.dtool`) {
		t.Fatalf("Dir = %q, want %q", got.Dir, filepath.Join(homeDir, `.dtool`))
	}
	if got.DBName != `memory.db` {
		t.Fatalf("DBName = %q, want %q", got.DBName, `memory.db`)
	}
	if got.DBPath != filepath.Join(homeDir, `.dtool`, `memory.db`) {
		t.Fatalf("DBPath = %q, want %q", got.DBPath, filepath.Join(homeDir, `.dtool`, `memory.db`))
	}
	if got.GitRepoEnabled {
		t.Fatalf("GitRepoEnabled = %v, want false", got.GitRepoEnabled)
	}
}

type fakeMemoryGitSyncer struct {
	isRepo       bool
	hasChanges   bool
	pullErr      error
	isGitRepoErr error
	hasChangeErr error
	addErr       error
	commitErr    error
	pushErr      error
	addCount     int
	commitCount  int
	pushCount    int
}

func (h *fakeMemoryGitSyncer) IsGitRepo(string) (bool, error) {
	return h.isRepo, h.isGitRepoErr
}

func (h *fakeMemoryGitSyncer) Pull(string) error {
	return h.pullErr
}

func (h *fakeMemoryGitSyncer) HasFileChanges(string, string) (bool, error) {
	return h.hasChanges, h.hasChangeErr
}

func (h *fakeMemoryGitSyncer) AddFile(string, string) error {
	h.addCount++
	return h.addErr
}

func (h *fakeMemoryGitSyncer) Commit(string, string, string) error {
	h.commitCount++
	return h.commitErr
}

func (h *fakeMemoryGitSyncer) Push(string) error {
	h.pushCount++
	return h.pushErr
}

func TestPrepareMemoryStoreReturnsPullErrorWhenGitPullFails(t *testing.T) {
	oldViper := component.ConfigViper
	oldPrepared := preparedMemoryStore
	oldFactory := newMemoryGitFactory
	t.Cleanup(func() {
		component.ConfigViper = oldViper
		preparedMemoryStore = oldPrepared
		newMemoryGitFactory = oldFactory
	})

	v := newTestMemoryConfigViper()
	v.Set(`base.memoryDbPath`, t.TempDir())
	v.Set(`base.memoryDbFileName`, `memory.db`)
	v.Set(`base.memoryDbIsGitRepo`, true)
	component.ConfigViper = v

	fakeGit := &fakeMemoryGitSyncer{
		isRepo:  true,
		pullErr: errors.New(`pull failed`),
	}
	newMemoryGitFactory = func() memoryGitSyncer {
		return fakeGit
	}

	err := PrepareMemoryStore()
	if err == nil {
		t.Fatalf("PrepareMemoryStore() error = nil, want error")
	}
	if err.Error() != `拉取记忆目录失败 pull failed` {
		t.Fatalf("PrepareMemoryStore() error = %q, want %q", err.Error(), `拉取记忆目录失败 pull failed`)
	}
}

func newTestMemoryConfigViper() *viper.Viper {
	return viper.New()
}

func TestSyncMemoryDBFilePushesChangedFile(t *testing.T) {
	fakeGit := &fakeMemoryGitSyncer{hasChanges: true}

	changed, err := SyncMemoryDBFile(common.MemoryConfig{
		Dir:       `C:\repo`,
		DBName:    `memory.db`,
		DBPath:    `C:\repo\memory.db`,
		IsGitRepo: true,
	}, fakeGit)
	if err != nil {
		t.Fatalf("SyncMemoryDBFile() error = %v", err)
	}
	if !changed {
		t.Fatalf("changed = false, want true")
	}
	if fakeGit.addCount != 1 || fakeGit.commitCount != 1 || fakeGit.pushCount != 1 {
		t.Fatalf("sync counts = add:%d commit:%d push:%d, want all 1", fakeGit.addCount, fakeGit.commitCount, fakeGit.pushCount)
	}
}

func TestSyncMemoryDBFileSkipsWhenNoChanges(t *testing.T) {
	fakeGit := &fakeMemoryGitSyncer{hasChanges: false}

	changed, err := SyncMemoryDBFile(common.MemoryConfig{
		Dir:       `C:\repo`,
		DBName:    `memory.db`,
		DBPath:    `C:\repo\memory.db`,
		IsGitRepo: true,
	}, fakeGit)
	if err != nil {
		t.Fatalf("SyncMemoryDBFile() error = %v", err)
	}
	if changed {
		t.Fatalf("changed = true, want false")
	}
	if fakeGit.addCount != 0 || fakeGit.commitCount != 0 || fakeGit.pushCount != 0 {
		t.Fatalf("sync counts = add:%d commit:%d push:%d, want all 0", fakeGit.addCount, fakeGit.commitCount, fakeGit.pushCount)
	}
}
