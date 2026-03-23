# 首页命令待执行队列 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为首页命令输入框增加待执行命令队列，使执行中的新命令能够按原始文本入队、可删除，并在前一个任务完成后自动串行执行。

**Architecture:** 在 `Dashboard.vue` 中新增最小化的待执行队列状态和调度逻辑，并将队列操作抽到独立工具模块，方便用轻量 Node 测试覆盖入队、出队、删除和顺序执行的关键语义。页面层仅负责展示队列和触发调度，不改变现有命令解析主流程。

**Tech Stack:** Vue 3、现有 `Dashboard.vue` 命令系统、Node 原生测试脚本

---

### Task 1: 新增命令队列纯函数模块

**Files:**
- Create: `web/src/utils/dashboard_command_queue.js`
- Test: `web/scripts/dashboard_command_queue.test.cjs`

**Step 1: Write the failing test**

在 `web/scripts/dashboard_command_queue.test.cjs` 中新增测试，断言：

- 新建队列项时会保留原始命令文本
- 入队后顺序保持 FIFO
- 删除指定 `id` 时只移除目标项

**Step 2: Run test to verify it fails**

Run: `node web/scripts/dashboard_command_queue.test.cjs`

Expected: FAIL，提示模块不存在或导出缺失

**Step 3: Write minimal implementation**

在 `web/src/utils/dashboard_command_queue.js` 中新增：

- 队列项字段常量
- `createPendingCommandItem`
- `enqueuePendingCommand`
- `dequeuePendingCommand`
- `removePendingCommandById`

并为关键判断补充中文注释。

**Step 4: Run test to verify it passes**

Run: `node web/scripts/dashboard_command_queue.test.cjs`

Expected: PASS

### Task 2: 补自动串行执行语义测试

**Files:**
- Modify: `web/scripts/dashboard_command_queue.test.cjs`
- Reference: `web/src/components/Dashboard.vue`

**Step 1: Write the failing test**

补测试覆盖：

- 连续入队两条命令时，出队顺序与入队顺序一致
- 删除队首后，下一条命令成为新的队首

**Step 2: Run test to verify it fails**

Run: `node web/scripts/dashboard_command_queue.test.cjs`

Expected: FAIL，提示缺少对应行为

**Step 3: Write minimal implementation**

仅在纯函数模块中补齐必要逻辑，不修改页面代码。

**Step 4: Run test to verify it passes**

Run: `node web/scripts/dashboard_command_queue.test.cjs`

Expected: PASS

### Task 3: 接入 Dashboard 执行中入队逻辑

**Files:**
- Modify: `web/src/components/Dashboard.vue`
- Test: `node web/scripts/dashboard_command_queue.test.cjs`

**Step 1: Write the failing test**

先确保现有纯函数测试已覆盖页面将使用的队列语义，然后人工校验 `Dashboard.vue` 现状不具备入队能力。

**Step 2: Run test to verify baseline**

Run: `node web/scripts/dashboard_command_queue.test.cjs`

Expected: PASS

**Step 3: Write minimal implementation**

在 `Dashboard.vue` 中：

- 新增待执行队列常量和响应式状态
- 新增“当前执行中时改为入队”的判断
- 新增入队提示文案

**Step 4: Run verification**

Run: `node web/scripts/dashboard_command_queue.test.cjs`

Expected: PASS

### Task 4: 接入完成后自动执行下一条

**Files:**
- Modify: `web/src/components/Dashboard.vue`
- Test: `node web/scripts/dashboard_command_queue.test.cjs`

**Step 1: Write the failing test**

补充或确认纯函数测试可以表达“取队首并继续处理”的语义。

**Step 2: Run test to verify baseline**

Run: `node web/scripts/dashboard_command_queue.test.cjs`

Expected: PASS

**Step 3: Write minimal implementation**

在 `finishExecution` 中增加：

- 若队列非空，取出队首
- 将命令文本重新注入现有解析执行链路
- 避免与当前 `isExecuting` 状态冲突

**Step 4: Run verification**

Run: `node web/scripts/dashboard_command_queue.test.cjs`

Expected: PASS

### Task 5: 增加右侧待执行列表与删除操作

**Files:**
- Modify: `web/src/components/Dashboard.vue`
- Test: `node web/scripts/dashboard_command_queue.test.cjs`

**Step 1: Write the failing test**

用人工对照方式确认模板尚无待执行列表区域。

**Step 2: Run test to verify baseline**

Run: `node web/scripts/dashboard_command_queue.test.cjs`

Expected: PASS

**Step 3: Write minimal implementation**

在 `Dashboard.vue` 中：

- 在输入框右侧增加待执行命令列表
- 展示数量和命令文本
- 支持删除单条待执行命令
- 为关键判断补充中文注释

**Step 4: Run verification**

Run: `node web/scripts/dashboard_command_queue.test.cjs`

Expected: PASS

### Task 6: 做基础回归验证

**Files:**
- Modify: `web/src/components/Dashboard.vue`
- Modify: `web/src/utils/dashboard_command_queue.js`
- Modify: `web/scripts/dashboard_command_queue.test.cjs`

**Step 1: Verify queue behavior**

Run: `node web/scripts/dashboard_command_queue.test.cjs`

Expected: PASS

**Step 2: Verify build-level syntax**

Run: `npm --prefix web run lint -- web/src/components/Dashboard.vue web/src/utils/dashboard_command_queue.js`

Expected: 若 CLI 不支持按文件 lint，则至少确认命令输出；禁止使用 `--fix`

**Step 3: Adjust only if needed**

仅修复本次改动引入的语法或明显格式问题，不扩大重构范围。
