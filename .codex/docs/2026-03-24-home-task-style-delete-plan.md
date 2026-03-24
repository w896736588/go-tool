# Home Task Style Delete Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make the home task panel visually align with the command shortcut screen, allow wheel page switching in the right-side blank area, and add irreversible task deletion.

**Architecture:** Keep the existing home task panel structure and data flow, but tighten the visual hierarchy in `Home.vue` to match the command shortcut area. Add task deletion through the existing home-task API stack by introducing a new delete endpoint from frontend utility to controller and sqlite access layer. Preserve the current archive/status flows and protect the wheel-switch behavior with focused regression tests.

**Tech Stack:** Vue 3, Element Plus, CommonJS utility tests with Node.js, Go + Gin, SQLite

---

### Task 1: Lock the wheel-switch regression with a failing test

**Files:**
- Modify: `web/scripts/home_dashboard_wheel.test.cjs`
- Test: `web/scripts/home_dashboard_wheel.test.cjs`

**Step 1: Write the failing test**

Add an assertion that a wheel event from the right-side blank area, represented by a non-scrollable child within the dashboard stage boundary, returns `false` from `shouldBlockHomeDashboardPageSwitch`.

**Step 2: Run test to verify it fails**

Run: `node web/scripts/home_dashboard_wheel.test.cjs`
Expected: FAIL on the blank-area assertion before the fix.

**Step 3: Write minimal implementation**

Keep `shouldBlockHomeDashboardPageSwitch` from blocking when no scrollable ancestor is found before the dashboard stage boundary.

**Step 4: Run test to verify it passes**

Run: `node web/scripts/home_dashboard_wheel.test.cjs`
Expected: PASS

### Task 2: Add a failing backend test for irreversible task deletion

**Files:**
- Modify: `internal/app/dtool/common/home_task_test.go`
- Modify: `internal/app/dtool/common/home_task.go`
- Test: `internal/app/dtool/common/home_task_test.go`

**Step 1: Write the failing test**

Add a test that creates a task, deletes it, verifies the delete call succeeds, and confirms `HomeTaskRow` no longer returns the task.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/dtool/common -run TestHomeTaskDeleteRemovesTask`
Expected: FAIL because `HomeTaskDelete` does not exist yet.

**Step 3: Write minimal implementation**

Add `HomeTaskDelete(id int) error` in the sqlite layer with parameter validation, existence check, and actual delete against `tbl_home_task`.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/dtool/common -run TestHomeTaskDeleteRemovesTask`
Expected: PASS

### Task 3: Expose task deletion through controller, request struct, router, and frontend API

**Files:**
- Modify: `internal/app/dtool/struct/home_task.go`
- Modify: `internal/app/dtool/controller/home_task.go`
- Modify: `internal/app/dtool/router.go`
- Modify: `web/src/utils/base/home_task.js`

**Step 1: Write the failing integration expectation**

Define the exact request shape as `{ id }` and wire a dedicated `HomeTaskDelete` API function on both frontend and backend.

**Step 2: Run focused verification**

Run: `go test ./internal/app/dtool/common -run TestHomeTaskDeleteRemovesTask`
Expected: still PASS after the API-layer wiring because the common-layer contract remains valid.

**Step 3: Write minimal implementation**

Add request struct, controller handler, router registration, and frontend utility wrapper without changing unrelated task APIs.

**Step 4: Run focused verification**

Run: `go test ./internal/app/dtool/common -run TestHomeTaskDeleteRemovesTask`
Expected: PASS

### Task 4: Add the delete action and visual alignment in the home task panel

**Files:**
- Modify: `web/src/components/Home.vue`

**Step 1: Write the failing UI expectation**

Add a frontend regression test only if the project already has a lightweight harness for this component; otherwise encode the behavior through the existing utility test and manual verification checklist, then implement the smallest DOM/style change set.

**Step 2: Run current focused checks**

Run: `node web/scripts/home_dashboard_wheel.test.cjs`
Expected: PASS before UI edits, providing a baseline for the wheel behavior.

**Step 3: Write minimal implementation**

Update the task panel header, toolbar, card spacing, and action area to visually align with the command shortcut screen. Add a danger-style delete entry with confirmation, call the new delete API, refresh both task lists, and keep archive/edit/status actions intact.

**Step 4: Run focused verification**

Run: `node web/scripts/home_dashboard_wheel.test.cjs`
Expected: PASS

### Task 5: Run final verification

**Files:**
- Verify: `web/scripts/home_dashboard_wheel.test.cjs`
- Verify: `internal/app/dtool/common/home_task_test.go`

**Step 1: Run frontend regression test**

Run: `node web/scripts/home_dashboard_wheel.test.cjs`
Expected: PASS

**Step 2: Run backend task tests**

Run: `go test ./internal/app/dtool/common -run HomeTask`
Expected: PASS

**Step 3: Manual verification checklist**

1. On `/首页`, hover inside the right-side blank area within about 200px of the right edge and scroll down/up; page switching still works.
2. In the command output area, internal scrolling still does not trigger whole-page switching until the scroll hits the edge.
3. The task panel title area, toolbar spacing, and action hierarchy feel visually aligned with the command shortcut section.
4. Deleting an active or archived task shows a destructive confirmation and removes the task after success.
