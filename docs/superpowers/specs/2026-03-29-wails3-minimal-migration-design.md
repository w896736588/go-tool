# Wails 3 Minimal Migration Design

**Date:** 2026-03-29

**Goal:** Upgrade the desktop client from `Wails v2` to `Wails v3` while preserving the current browser mode and desktop client behavior.

## Scope

- Keep the existing browser mode unchanged.
- Keep the existing desktop client entry point, startup parameter, and packaging task names unchanged.
- Preserve the current desktop flow:
  - Start the desktop shell.
  - Load embedded frontend assets as the initial page.
  - Start the local backend service after the desktop frontend is ready.
  - Wait for the local backend port to become reachable.
  - Redirect the desktop window to the local backend URL.

## Non-Goals

- Do not migrate the application to Wails 3 service/bindings architecture.
- Do not replace the local backend URL redirection flow.
- Do not redesign the frontend or change the browser-mode product behavior.
- Do not rename existing task commands or packaging outputs unless required by Wails 3 API changes.

## Current State

The current desktop entry uses `github.com/wailsapp/wails/v2` in [cmd/dtool_wails/main.go](/C:/work/frog/dev_tool_master/cmd/dtool_wails/main.go) and uses `github.com/wailsapp/wails/v2/pkg/runtime` in [internal/app/dtool/wailsapp/app.go](/C:/work/frog/dev_tool_master/internal/app/dtool/wailsapp/app.go).

The desktop lifecycle logic is:

1. Save runtime context during startup.
2. After DOM ready, asynchronously boot the backend.
3. Poll the backend port until reachable.
4. Execute browser-side JavaScript to redirect the window to the backend URL.
5. Stop backend resources during shutdown.

## Target Design

### Entry Layer

The desktop entry will move from the Wails 2 `wails.Run(&options.App{...})` pattern to the Wails 3 `application.New(...)` and window-construction pattern.

Responsibilities:

- Create the Wails 3 application instance.
- Create a single desktop window with the existing title, size, and minimum size.
- Configure embedded frontend assets for the initial page.
- Wire startup, DOM-ready, and shutdown lifecycle hooks to the existing backend boot logic.

### Desktop Lifecycle Layer

The `DesktopApp` type will keep the same high-level responsibilities, but its runtime coupling will change:

- Remove direct dependency on the Wails 2 runtime package.
- Store the Wails 3 application or window handle needed for logging and JavaScript execution.
- Keep backend boot guarded by `sync.Once`.
- Keep port polling behavior unchanged.
- Keep shutdown behavior unchanged.

### Redirect Behavior

The desktop client will continue to redirect to the locally served backend URL after the backend becomes reachable.

This is the key compatibility rule for the migration:

- Wails 3 is only the desktop shell upgrade.
- The actual application content and backend integration flow remain unchanged.

## File-Level Impact

### Files to Modify

- [go.mod](/C:/work/frog/dev_tool_master/go.mod)
  - Replace `github.com/wailsapp/wails/v2` dependency with the Wails 3 dependency set needed by the current entry pattern.
- [go.sum](/C:/work/frog/dev_tool_master/go.sum)
  - Update dependency checksums after the migration.
- [cmd/dtool_wails/main.go](/C:/work/frog/dev_tool_master/cmd/dtool_wails/main.go)
  - Rewrite the entry to Wails 3 application/window APIs.
- [internal/app/dtool/wailsapp/app.go](/C:/work/frog/dev_tool_master/internal/app/dtool/wailsapp/app.go)
  - Replace Wails 2 runtime calls with Wails 3 application/window operations.

### Files Expected To Stay Stable

- [Taskfile.yml](/C:/work/frog/dev_tool_master/Taskfile.yml)
- [script/package_windows.ps1](/C:/work/frog/dev_tool_master/script/package_windows.ps1)
- Browser-mode backend files under `cmd/dtool` and `internal/app/dtool`
- Frontend source and build configuration under `web/`

## Error Handling

- If the local backend port does not become ready within the timeout, log a desktop-side error and do not redirect.
- If the JavaScript redirect fails, surface the Wails 3 error through the application logger where possible.
- Shutdown must still call `dtool.Stop()` even if startup did not complete successfully.

## Compatibility Constraints

- The `--ConfigFile` startup parameter must remain supported.
- Existing package tasks such as `task package-windows` must keep working after the migration.
- Windows desktop compilation is the primary verification target because it is the currently failing and packaged path.
- Linux and macOS desktop compile paths should also be validated if the Wails 3 API surface requires cross-platform adjustments.

## Testing Strategy

### Build Verification

- Verify the desktop target compiles with Wails 3.
- Verify existing web-mode build and packaging tasks still run.
- Verify `task package-windows` succeeds end-to-end.

### Behavior Verification

- Confirm the desktop app still boots the backend once.
- Confirm the desktop app still redirects to the local backend URL after the backend is ready.
- Confirm desktop shutdown still stops backend resources.

## Risks

- Wails 3 is currently alpha, so API adjustments may be needed during migration.
- Window lifecycle hook names and JS execution APIs may differ from Wails 2 and require small adapter changes.
- Asset serving configuration may differ between Wails 2 and Wails 3, which could affect the initial embedded page if not mapped correctly.

## Recommended Implementation Path

1. Upgrade dependencies to Wails 3.
2. Port the desktop entry from Wails 2 app options to Wails 3 application/window creation.
3. Adapt `DesktopApp` to use Wails 3 handles instead of Wails 2 runtime context.
4. Rebuild the desktop binary.
5. Re-run full Windows packaging.

