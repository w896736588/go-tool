# Home Task AI Work Report Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a one-click AI daily work report action above the home task list that uses configured model/prompt, summarizes active tasks on the backend, and saves the result into memory with the `工作日报` tag.

**Architecture:** Extend the existing memory settings page with a dedicated daily-report model and prompt, add one homepage trigger button, and implement a backend controller path that loads active home tasks, builds a structured prompt, calls the configured LLM, and writes the generated Markdown into the memory store. Reuse the existing global config, AI model lookup, chat invocation, and memory save/sync flows to minimize new surface area.

**Tech Stack:** Vue 3, Element Plus, Go, Gin, SQLite global config, existing memory fragment store, existing AI model/provider pipeline

---

### Task 1: Add failing backend tests for report config and prompt assembly

**Files:**
- Modify: `D:/go/cache_manager_api/internal/app/dtool/controller/memory_fragment_test.go`
- Create: `D:/go/cache_manager_api/internal/app/dtool/controller/home_task_report_test.go`

**Step 1: Write the failing tests**

Add tests for:

- default daily report prompt fallback returns non-empty text
- report title builder returns `工作日报 YYYY-MM-DD`
- task snapshot builder includes name, status, start time, last operated time, and remark
- empty task list returns an explicit error

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/app/dtool/controller -run "Test(BuildHomeTaskDailyReport|DefaultHomeTaskDailyReport)" -count=1`

Expected: FAIL because the new builder/helper functions do not exist yet.

**Step 3: Commit**

```bash
git add internal/app/dtool/controller/home_task_report_test.go internal/app/dtool/controller/memory_fragment_test.go
git commit -m "test: cover home task daily report helpers"
```

### Task 2: Add backend configuration constants and config read/write support

**Files:**
- Modify: `D:/go/cache_manager_api/internal/app/dtool/define/global.go`
- Modify: `D:/go/cache_manager_api/internal/app/dtool/controller/set.go`

**Step 1: Write the failing test**

Add a test covering:

- `SetMemoryConfigGet` returns the new daily report keys
- `SetMemoryConfigSave` validates selected model type as `llm`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/dtool/controller -run "TestSetMemoryConfig" -count=1`

Expected: FAIL because the new config fields are not exposed or validated yet.

**Step 3: Write minimal implementation**

Implement:

- new global keys for daily report model and prompt
- `SetMemoryConfigGet` response fields:
  - `home_task_daily_report_model_id`
  - `home_task_daily_report_prompt`
- `SetMemoryConfigSave` persistence and `llm` validation for the new model id
- default prompt fallback when the submitted prompt is blank

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/dtool/controller -run "TestSetMemoryConfig" -count=1`

Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/dtool/define/global.go internal/app/dtool/controller/set.go
git commit -m "feat: add home task daily report config"
```

### Task 3: Implement backend report generation and memory persistence

**Files:**
- Modify: `D:/go/cache_manager_api/internal/app/dtool/controller/memory_fragment.go`
- Modify: `D:/go/cache_manager_api/internal/app/dtool/router.go`
- Modify: `D:/go/cache_manager_api/internal/app/dtool/common/home_task.go`
- Modify: `D:/go/cache_manager_api/internal/app/dtool/controller/home_task.go`
- Modify: `D:/go/cache_manager_api/internal/app/dtool/struct/home_task.go`

**Step 1: Write the failing test**

Add controller/helper tests covering:

- generation aborts when memory is not configured
- generation aborts when no active tasks exist
- generation uses configured model and prompt
- saved memory fragment title and tags are correct

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/dtool/controller -run "TestHomeTaskDailyReportGenerate" -count=1`

Expected: FAIL because the endpoint and helper pipeline do not exist yet.

**Step 3: Write minimal implementation**

Implement:

- a new backend entry such as `/api/HomeTaskDailyReportGenerate`
- a request-free controller action that:
  - ensures memory is configured
  - loads active home tasks
  - loads daily report config
  - builds the user prompt
  - calls `InfoCrawlChatByModel`
  - strips markdown fence if needed
  - saves the result through the memory DB with title `工作日报 YYYY-MM-DD` and tag `工作日报`
  - triggers `common.MemoryRuntime.ScheduleSync()`
- helper functions with comments for:
  - default prompt
  - config lookup
  - report title
  - task snapshot formatting
  - user prompt assembly

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/dtool/controller -run "TestHomeTaskDailyReportGenerate|Test(BuildHomeTaskDailyReport|DefaultHomeTaskDailyReport)" -count=1`

Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/dtool/controller/memory_fragment.go internal/app/dtool/router.go internal/app/dtool/common/home_task.go internal/app/dtool/controller/home_task.go internal/app/dtool/struct/home_task.go
git commit -m "feat: add home task daily report generation"
```

### Task 4: Expose the new config in the settings UI

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/components/set/memory.vue`
- Modify: `D:/go/cache_manager_api/web/src/utils/base/git_set.js`

**Step 1: Write the failing UI check**

Add a minimal component-level check if test infra exists; otherwise document the manual verification target:

- memory settings page loads daily report model and prompt
- save submits both new fields

**Step 2: Run the relevant verification**

If UI tests exist, run that single test.

If no UI tests exist, use build verification later and confirm code paths manually.

**Step 3: Write minimal implementation**

Implement:

- new “工作日报 AI” section in the memory settings page
- model selector bound to `home_task_daily_report_model_id`
- prompt textarea bound to `home_task_daily_report_prompt`
- default prompt constant
- load/save wiring through `MemoryConfigGet` / `MemoryConfigSave`

**Step 4: Verify**

Run: `npm run prod`

Expected: PASS build with no new errors.

**Step 5: Commit**

```bash
git add web/src/components/set/memory.vue web/src/utils/base/git_set.js
git commit -m "feat: expose home task daily report settings"
```

### Task 5: Add the one-click homepage trigger

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/components/Home.vue`
- Modify: `D:/go/cache_manager_api/web/src/utils/base/home_task.js`

**Step 1: Write the failing UI check**

Define the verification target:

- homepage toolbar shows `AI 生成工作日报`
- clicking it calls the new API
- button enters loading state during the request
- success and failure notifications are displayed correctly

**Step 2: Run the relevant verification**

If no targeted UI test infra exists, rely on production build plus manual reasoning around the small change surface.

**Step 3: Write minimal implementation**

Implement:

- new toolbar button and loading state constants/data
- new frontend API wrapper for `/api/HomeTaskDailyReportGenerate`
- success callback message
- failure callback message
- duplicate-click guard during generation

**Step 4: Verify**

Run: `npm run prod`

Expected: PASS build with no new errors.

**Step 5: Commit**

```bash
git add web/src/components/Home.vue web/src/utils/base/home_task.js
git commit -m "feat: trigger home task ai daily report from homepage"
```

### Task 6: Run final verification

**Files:**
- Modify: none

**Step 1: Run backend verification**

Run: `go test ./internal/app/dtool/controller -count=1`

Expected: PASS

**Step 2: Run frontend verification**

Run: `npm run prod`

Expected: PASS build with only pre-existing warnings.

**Step 3: Inspect final diff**

Run: `git diff --stat`

Expected: only the planned files changed.

**Step 4: Commit**

```bash
git add .
git commit -m "feat: add ai daily work report for home tasks"
```
