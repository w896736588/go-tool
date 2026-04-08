package controller

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeGitBranchShell struct {
	output    string
	run       func(command string) (string, error)
	closeFunc func()
}

func (f *fakeGitBranchShell) RunCommandOnce(command string) (string, error) {
	if f.run != nil {
		return f.run(command)
	}
	return f.output, nil
}

func (f *fakeGitBranchShell) Close() {
	if f.closeFunc != nil {
		f.closeFunc()
	}
}

func TestRunGitGroupBranchQueriesUsesFreshConnectionsAndReleasesThem(t *testing.T) {
	gitList := make([]map[string]any, 0, 7)
	for index := 0; index < 7; index++ {
		gitList = append(gitList, map[string]any{
			`name`:      fmt.Sprintf("repo-%d", index),
			`code_path`: fmt.Sprintf("/var/www/repo-%d", index),
		})
	}

	var mu sync.Mutex
	createdIDs := make([]string, 0, len(gitList))
	releasedIDs := make([]string, 0, len(gitList))
	closedCount := 0

	results := runGitGroupBranchQueries(
		gitList,
		map[string]any{`id`: `ssh-1`},
		func(string) {},
		func(sshConfig map[string]any) (gitBranchRunner, error) {
			mu.Lock()
			createdIDs = append(createdIDs, fmt.Sprintf("runner-%d", len(createdIDs)+1))
			mu.Unlock()
			return &fakeGitBranchShell{
				output: "当前分支：\nfeature/test\n远程分支：\norigin/feature/test\n",
				closeFunc: func() {
					mu.Lock()
					closedCount++
					mu.Unlock()
				},
			}, nil
		},
		func() {
			mu.Lock()
			releasedIDs = append(releasedIDs, fmt.Sprintf("released-%d", len(releasedIDs)+1))
			mu.Unlock()
		},
	)

	if len(results) != len(gitList) {
		t.Fatalf("len(results) = %d, want %d", len(results), len(gitList))
	}
	if len(createdIDs) != len(gitList) {
		t.Fatalf("len(createdIDs) = %d, want %d", len(createdIDs), len(gitList))
	}
	if len(releasedIDs) != len(gitList) {
		t.Fatalf("len(releasedIDs) = %d, want %d", len(releasedIDs), len(gitList))
	}
	if closedCount != len(gitList) {
		t.Fatalf("closedCount = %d, want %d", closedCount, len(gitList))
	}

	createdSet := make(map[string]struct{}, len(createdIDs))
	for _, shellClientId := range createdIDs {
		createdSet[shellClientId] = struct{}{}
	}
	if len(createdSet) != len(gitList) {
		t.Fatalf("unique created shell ids = %d, want %d", len(createdSet), len(gitList))
	}

	releasedSet := make(map[string]struct{}, len(releasedIDs))
	for _, shellClientId := range releasedIDs {
		releasedSet[shellClientId] = struct{}{}
	}
	if len(releasedSet) != len(gitList) {
		t.Fatalf("unique released shell ids = %d, want %d", len(releasedSet), len(gitList))
	}

	for index, result := range results {
		if result[`local_branch`] != `feature/test` {
			t.Fatalf("results[%d].local_branch = %v, want feature/test", index, result[`local_branch`])
		}
		if result[`remote_branch`] != `origin/feature/test` {
			t.Fatalf("results[%d].remote_branch = %v, want origin/feature/test", index, result[`remote_branch`])
		}
	}
}

func TestRunGitGroupBranchQueriesLimitsConcurrencyToFive(t *testing.T) {
	gitList := make([]map[string]any, 0, 8)
	for index := 0; index < 8; index++ {
		gitList = append(gitList, map[string]any{
			`name`:      fmt.Sprintf("repo-%d", index),
			`code_path`: fmt.Sprintf("/var/www/repo-%d", index),
		})
	}

	var currentConcurrent int32
	var maxConcurrent int32

	_ = runGitGroupBranchQueries(
		gitList,
		map[string]any{`id`: `ssh-1`},
		func(string) {},
		func(sshConfig map[string]any) (gitBranchRunner, error) {
			return &fakeGitBranchShell{
				run: func(command string) (string, error) {
					current := atomic.AddInt32(&currentConcurrent, 1)
					for {
						seen := atomic.LoadInt32(&maxConcurrent)
						if current <= seen || atomic.CompareAndSwapInt32(&maxConcurrent, seen, current) {
							break
						}
					}
					time.Sleep(50 * time.Millisecond)
					atomic.AddInt32(&currentConcurrent, -1)
					return "当前分支：\nmain\n远程分支：\norigin/main\n", nil
				},
			}, nil
		},
		func() {},
	)

	if maxConcurrent != gitGroupBranchListConcurrency {
		t.Fatalf("maxConcurrent = %d, want %d", maxConcurrent, gitGroupBranchListConcurrency)
	}
}

func TestRunGitGroupBranchQueriesDoesNotBindRawSSEStream(t *testing.T) {
	gitList := []map[string]any{
		{
			`name`:      `repo-1`,
			`code_path`: `/var/www/repo-1`,
		},
	}

	results := runGitGroupBranchQueries(
		gitList,
		map[string]any{`id`: `ssh-1`},
		func(string) {},
		func(sshConfig map[string]any) (gitBranchRunner, error) {
			return &fakeGitBranchShell{
				output: "当前分支：\nmain\n远程分支：\norigin/main\n",
			}, nil
		},
		func() {},
	)

	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0][`local_branch`] != `main` {
		t.Fatalf("local_branch = %v, want main", results[0][`local_branch`])
	}
}
