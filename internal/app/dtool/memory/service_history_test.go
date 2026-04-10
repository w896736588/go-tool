package memory

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMemoryFragmentHistoryListReadsGitHistory(t *testing.T) {
	root := t.TempDir()
	runGitCommand(t, root, `init`)
	runGitCommand(t, root, `config`, `user.name`, `Codex Test`)
	runGitCommand(t, root, `config`, `user.email`, `codex@example.com`)

	service := NewService(root)
	now := time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC)
	filePath := BuildFragmentPath(root, now, `demo-fragment`, false)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	firstContent := strings.TrimSpace(`---
title: 第一版标题
created_at: 2026-04-10T09:00:00Z
updated_at: 2026-04-10T09:00:00Z
---

# 第一版标题

第一版内容
`)
	if err := os.WriteFile(filePath, []byte(firstContent), 0o644); err != nil {
		t.Fatalf("WriteFile(first) error = %v", err)
	}
	runGitCommand(t, root, `add`, `--`, `.`)
	runGitCommand(t, root, `commit`, `-m`, `feat: add memory fragment`)

	secondContent := strings.TrimSpace(`---
title: 第二版标题
created_at: 2026-04-10T09:00:00Z
updated_at: 2026-04-10T10:00:00Z
---

# 第二版标题

第二版内容
`)
	if err := os.WriteFile(filePath, []byte(secondContent), 0o644); err != nil {
		t.Fatalf("WriteFile(second) error = %v", err)
	}
	runGitCommand(t, root, `add`, `--`, `.`)
	runGitCommand(t, root, `commit`, `-m`, `fix: update memory fragment`)

	service.upsert(Fragment{
		ID:        `demo-fragment`,
		Title:     `第二版标题`,
		CreatedAt: now,
		UpdatedAt: now.Add(time.Hour),
		FilePath:  filePath,
	})

	list, err := service.MemoryFragmentHistoryList(`demo-fragment`)
	if err != nil {
		t.Fatalf("MemoryFragmentHistoryList() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len(history) = %d, want 2", len(list))
	}
	if got := list[0][`change_desc`]; got != `fix: update memory fragment` {
		t.Fatalf("latest change_desc = %v, want update commit message", got)
	}
	if got := list[0][`title_old`]; got != `第一版标题` {
		t.Fatalf("latest title_old = %v, want 第一版标题", got)
	}
	if got := list[0][`title_new`]; got != `第二版标题` {
		t.Fatalf("latest title_new = %v, want 第二版标题", got)
	}
	if got := strings.TrimSpace(list[0][`content_old`].(string)); !strings.Contains(got, `第一版内容`) {
		t.Fatalf("latest content_old = %q, want first revision", got)
	}
	if got := strings.TrimSpace(list[0][`content_new`].(string)); !strings.Contains(got, `第二版内容`) {
		t.Fatalf("latest content_new = %q, want second revision", got)
	}
	if got := list[1][`change_desc`]; got != `feat: add memory fragment` {
		t.Fatalf("oldest change_desc = %v, want create commit message", got)
	}
	if got := list[1][`title_old`]; got != `` {
		t.Fatalf("create commit title_old = %v, want empty", got)
	}
	if got := list[1][`title_new`]; got != `第一版标题` {
		t.Fatalf("create commit title_new = %v, want 第一版标题", got)
	}
}

func TestMemoryFragmentHistoryListReadsGitHistoryFromRepoSubdir(t *testing.T) {
	repoRoot := t.TempDir()
	memoryRoot := filepath.Join(repoRoot, `memory`)
	if err := os.MkdirAll(memoryRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll(memoryRoot) error = %v", err)
	}
	runGitCommand(t, repoRoot, `init`)
	runGitCommand(t, repoRoot, `config`, `user.name`, `Codex Test`)
	runGitCommand(t, repoRoot, `config`, `user.email`, `codex@example.com`)

	service := NewService(memoryRoot)
	now := time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC)
	filePath := BuildFragmentPath(memoryRoot, now, `demo-fragment`, false)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	firstContent := strings.TrimSpace(`---
title: 子目录第一版
created_at: 2026-04-10T09:00:00Z
updated_at: 2026-04-10T09:00:00Z
---

# 子目录第一版

第一版内容
`)
	if err := os.WriteFile(filePath, []byte(firstContent), 0o644); err != nil {
		t.Fatalf("WriteFile(first) error = %v", err)
	}
	runGitCommand(t, repoRoot, `add`, `--`, `.`)
	runGitCommand(t, repoRoot, `commit`, `-m`, `feat: add nested memory fragment`)

	secondContent := strings.TrimSpace(`---
title: 子目录第二版
created_at: 2026-04-10T09:00:00Z
updated_at: 2026-04-10T10:00:00Z
---

# 子目录第二版

第二版内容
`)
	if err := os.WriteFile(filePath, []byte(secondContent), 0o644); err != nil {
		t.Fatalf("WriteFile(second) error = %v", err)
	}
	runGitCommand(t, repoRoot, `add`, `--`, `.`)
	runGitCommand(t, repoRoot, `commit`, `-m`, `fix: update nested memory fragment`)

	service.upsert(Fragment{
		ID:        `demo-fragment`,
		Title:     `子目录第二版`,
		CreatedAt: now,
		UpdatedAt: now.Add(time.Hour),
		FilePath:  filePath,
	})

	list, err := service.MemoryFragmentHistoryList(`demo-fragment`)
	if err != nil {
		t.Fatalf("MemoryFragmentHistoryList() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len(history) = %d, want 2", len(list))
	}
	if got := list[0][`change_desc`]; got != `fix: update nested memory fragment` {
		t.Fatalf("latest change_desc = %v, want nested update commit message", got)
	}
}

func runGitCommand(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(`git`, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output))
}
