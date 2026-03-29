# Wails 3 Minimal Migration Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the desktop client from Wails v2 to Wails v3 without changing the existing browser mode or the desktop redirect-to-local-backend behavior.

**Architecture:** Keep the current desktop lifecycle and local backend boot flow intact, but replace the Wails v2 app bootstrap and runtime APIs with the Wails v3 application/window model. Limit changes to dependency management, desktop entry wiring, and desktop lifecycle integration so existing packaging tasks and browser-mode code remain stable.

**Tech Stack:** Go 1.26, Wails v3 alpha CLI/runtime, embedded frontend assets, Task, PowerShell packaging script

---

## Chunk 1: Dependency And API Surface Migration

### Task 1: Update Wails module dependencies

**Files:**
- Modify: `C:/work/frog/dev_tool_master/go.mod`
- Modify: `C:/work/frog/dev_tool_master/go.sum`

- [ ] **Step 1: Inspect current Wails v2 dependency usage**

Run: `rg -n "wails/v2|wailsapp/wails" go.mod cmd internal`
Expected: Only desktop entry files and module dependency reference Wails v2.

- [ ] **Step 2: Update module dependency to Wails v3**

Change `github.com/wailsapp/wails/v2` to the Wails v3 module needed by the desktop entry.

- [ ] **Step 3: Download updated module graph**

Run: `go mod download`
Expected: `go.sum` updates with Wails v3 module checksums.

- [ ] **Step 4: Verify the module resolution**

Run: `go list -m github.com/wailsapp/wails/v3`
Expected: Reports the installed Wails v3 version.

### Task 2: Map Wails v2 lifecycle calls to Wails v3 equivalents

**Files:**
- Modify: `C:/work/frog/dev_tool_master/cmd/dtool_wails/main.go`
- Modify: `C:/work/frog/dev_tool_master/internal/app/dtool/wailsapp/app.go`

- [ ] **Step 1: Identify the current lifecycle and runtime touch points**

Run: `rg -n "Startup|DomReady|Shutdown|runtime\\.|WindowExecJS|LogErrorf" cmd/dtool_wails internal/app/dtool/wailsapp`
Expected: Exact list of calls that must be ported to Wails v3 APIs.

- [ ] **Step 2: Rewrite the desktop entry to Wails v3 application/window bootstrap**

Implement the minimal Wails v3 bootstrap in `cmd/dtool_wails/main.go`, preserving title, size, minimum size, embedded assets, and the `--ConfigFile` behavior.

- [ ] **Step 3: Adapt desktop lifecycle integration to Wails v3 handles**

Update `DesktopApp` so it stores the Wails v3 window or app handle needed for:
- desktop-side logging
- executing the redirect JavaScript
- preserving startup / DOM-ready / shutdown behavior

- [ ] **Step 4: Add concise bilingual comments to new key constants, methods, and key branching**

Follow repo guidance for Chinese and English comments on critical methods and key decisions.

## Chunk 2: Build And Packaging Verification

### Task 3: Verify desktop compilation with Wails v3

**Files:**
- Verify: `C:/work/frog/dev_tool_master/cmd/dtool_wails/main.go`
- Verify: `C:/work/frog/dev_tool_master/internal/app/dtool/wailsapp/app.go`

- [ ] **Step 1: Run the desktop build directly**

Run: `go build -tags production -ldflags "-s -w -H=windowsgui" -o build/dtool_wails.exe ./cmd/dtool_wails`
Expected: Build succeeds with no Wails v2 import errors.

- [ ] **Step 2: Fix any compile-time API mismatches one at a time**

If the build fails, patch only the specific Wails v3 API mismatch and re-run the same build command before making further changes.

### Task 4: Verify end-to-end Windows packaging still works

**Files:**
- Verify: `C:/work/frog/dev_tool_master/Taskfile.yml`
- Verify: `C:/work/frog/dev_tool_master/script/package_windows.ps1`

- [ ] **Step 1: Re-run Windows packaging**

Run: `task package-windows`
Expected: Frontend build, web build, desktop build, and packaging all succeed.

- [ ] **Step 2: Confirm the expected archive is produced**

Check: `C:/work/frog/dev_tool_master/build`
Expected: A fresh `dtool_release_windows_*.zip` package exists.

- [ ] **Step 3: Summarize residual warnings separately from failures**

Record non-blocking warnings such as frontend bundle size or Node engine warnings, but do not treat them as migration failures unless they change behavior.

