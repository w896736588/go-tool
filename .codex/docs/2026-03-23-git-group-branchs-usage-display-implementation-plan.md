# Git Group Branchs Usage Display Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 调整 `git group-branchs` 的“是否有人使用”列，让本地分支不存在或无人使用时统一返回 `-`。

**Architecture:** 在 `internal/app/dtool/controller/git.go` 内收敛“是否有人使用”列的展示规则，优先根据本地分支是否存在做短路判断，再保留远程分支存在和最近工作区属主回退逻辑。通过 `internal/app/dtool/controller/git_group_usage_test.go` 增补回归测试，保证 `-` 的新语义稳定。

**Tech Stack:** Go、标准库 `testing`

---

### Task 1: 调整 usage 展示规则

**Files:**
- Modify: `internal/app/dtool/controller/git_group_usage_test.go`
- Modify: `internal/app/dtool/controller/git.go`

**Step 1: Write the failing test**

在 `internal/app/dtool/controller/git_group_usage_test.go` 中补充和调整以下断言：
- 本地分支存在且远程分支存在时，返回“有人使用”
- 本地分支存在但无人使用时，返回 `-`
- 本地分支不存在时，直接返回 `-`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/dtool/controller -run "TestBuildBranchUsageDisplay|TestParseGitStatusEntries|TestParseRecentUsageOwners"`
Expected: FAIL，旧逻辑仍返回“无人使用”或未区分本地分支缺失。

**Step 3: Write minimal implementation**

在 `internal/app/dtool/controller/git.go` 中：
- 提取“有人使用”和 `-` 为常量，并增加注释
- 在 `queryBranchUsageInfo` 中优先判断 `LocalBranch`
- 调整 `buildBranchUsageDisplay` 参数，纳入 `localBranch`
- 将无人使用返回值统一为 `-`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/dtool/controller -run "TestBuildBranchUsageDisplay|TestParseGitStatusEntries|TestParseRecentUsageOwners"`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/dtool/controller/git.go internal/app/dtool/controller/git_group_usage_test.go .codex/docs/2026-03-23-git-group-branchs-usage-display-implementation-plan.md
git commit -m "fix: unify git group branch usage display"
```
