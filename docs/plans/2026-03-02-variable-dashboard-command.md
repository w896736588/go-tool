# Variable 快捷命令多步交互 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在首页 Dashboard 中把 `variable` 做成可执行的快捷命令，并支持输入框/选项的多步交互流程。

**Architecture:** 在 `commandConfig` 增加 `variable` 子命令定义；在 `Dashboard.vue` 引入 `variable_set` API，并维护一个单实例 `variableSession` 状态机。通过 `VariableRun` 与 `VariableSet` 驱动状态流转，配合动态候选项实现“步骤选择 + 参数输入 + 最终执行”。

**Tech Stack:** Vue 3 Composition API、Element Plus、现有 `web/src/utils/base/variable_set.js` 接口。

---

### Task 1: 扩展首页命令配置

**Files:**
- Modify: `web/src/config/commandConfig.js`

**Step 1: 新增 variable 子命令定义**
- 增加 `run/set/choose/exec/reset/cancel`。
- `run` 使用 `dynamicChildren: 'variableScriptList'` 和 `needTarget: true`。
- `choose` 使用 `dynamicChildren: 'variableOptionList'` 和 `needTarget: true`。
- `set` 使用 `needInput: true`。

**Step 2: 自检配置一致性**
- 确认 action 名称与 Dashboard 分发一致。

### Task 2: 在 Dashboard 中接入 variable 会话状态

**Files:**
- Modify: `web/src/components/Dashboard.vue`

**Step 1: 引入变量 API 并新增状态**
- `import variableSet from '@/utils/base/variable_set'`
- 新增 `variableSession` 状态：`active/variableId/variableName/runCmdId/replaceList/isRun/isFinish/currentForm`

**Step 2: 新增动态加载器**
- `loadVariableScriptList()`：调用 `VariableList`，缓存至 `dynamicDataCache['variableScriptList']`。
- `loadVariableOptionList()`：从 `variableSession.currentForm.Select.OptionList` 生成候选。

**Step 3: 动作分发与执行器**
- 在 `executeAction` 增加 `variableRun/variableSet/variableChoose/variableExec/variableReset/variableCancel` 分支。
- 实现：
  - `executeVariableRunAction(stack)`
  - `executeVariableSessionAction(action, stack, inputValue)`
  - `handleVariableFlowResponse(response)`
  - `resetVariableSession(reason)`

### Task 3: 交互提示与错误处理

**Files:**
- Modify: `web/src/components/Dashboard.vue`

**Step 1: 统一提示文案（中文）**
- 按 `RunStatus` 输出下一步建议命令。

**Step 2: 增加防错分支**
- 无会话、步骤类型不匹配、选项不存在、后端失败等。

### Task 4: 验证

**Files:**
- Modify: 无

**Step 1: 静态检查**
- 运行：`npm run lint`（若项目支持）

**Step 2: 手工回归清单**
- `variable run <脚本>` -> `set/choose` -> `exec`。
- 错误路径：无会话直接 `set/choose/exec`。
