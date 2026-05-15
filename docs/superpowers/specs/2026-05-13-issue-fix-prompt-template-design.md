# 问题修改提示词模板 - 设计方案

## 概述

在现有工作流提示词模板体系中新增"问题修改提示词模板"，并在任务工作流页面提供独立的快捷按钮，用于快速生成包含改动要求和已解析提示词的完整文本。

## 一、后端改动

### 1.1 新增常量

**文件**: `internal/app/dtool/define/home_task_config.go`

新增一行常量：
```go
HomeTaskConfigPromptIssueFix = `home_task_prompt_issue_fix`
```

### 1.2 数据库迁移

**新建文件**: `internal/app/dtool/database/2026/05/20260513100000-task_workflow_issue_fix.sql`

给 `tbl_task_workflow` 新增 `prompt_issue_fix` 列（TEXT，默认空字符串）。

### 1.3 结构体

**文件**: `internal/app/dtool/struct/task_workflow.go`

`TaskWorkflow` 结构体新增字段 `PromptIssueFix`（json: `prompt_issue_fix`）。

### 1.4 DB 层

**文件**: `internal/app/dtool/common/task_workflow.go`

- `TaskWorkflowCreateOrGetByHomeTaskID`: 不修改（新 workflow 创建时只填充初始字段）
- `TaskWorkflowUpdatePrompts` / `taskWorkflowAppendSetClauses`: 加入 `prompt_issue_fix` 字段的读写

### 1.5 Controller

**文件**: `internal/app/dtool/controller/task_workflow.go`

- `resolveTaskWorkflowPrompts`: 加入 `issue_fix` 的模板读取和占位符解析
- 新增方法 `TaskWorkflowIssueFixResolve`: 提供 API `/api/task/workflow/issue-fix/resolve`（POST）
  - 入参: `workflow_id`
  - 逻辑: 读取 `home_task_prompt_issue_fix` 模板 → 构建占位符映射 → 替换占位符 → 返回解析后的文本
  - 响应: `{ "prompt": "已解析的提示词文本" }`

### 1.6 路由

**文件**: `internal/app/dtool/router.go`

新增路由：`POST /api/task/workflow/issue-fix/resolve`

## 二、前端改动

### 2.1 设置页新增模板

**文件**: `web/src/components/set/home_task_report.vue`

- 在"工作流提示词模板"的 `<el-tabs>` 中新增子 tab `issue_fix`，标签为"问题修改提示词"
- 绑定 `form.home_task_prompt_issue_fix`
- 在 data() 中新增字段 `home_task_prompt_issue_fix: ''`
- 在 load 方法中加入从 `response.Data.home_task_prompt_issue_fix` 读取
- 在 save 方法中加入 `home_task_prompt_issue_fix: this.form.home_task_prompt_issue_fix`

### 2.2 工作流页新增按钮和弹窗

**文件**: `web/src/components/TaskWorkflow.vue`

在 header 的 actions 区域新增按钮"问题修改提示词"。点击后弹窗：

- **弹窗上方**: `<el-input type="textarea">` 输入框，placeholder 为"请描述需要修改的问题"
- **弹窗下方**: `<MdEditor>`（只读预览模式）展示组合后的完整内容
  - 内容 = 用户输入 + 换行 + 调用 API 获取的已解析模板
- **弹窗底部**: `<el-button>` "复制到剪贴板"，复制完整 MD 内容

弹窗实现为独立组件或在 `TaskWorkflow.vue` 内以 `<el-dialog>` 实现。

交互流程：
1. 点击按钮 → 弹窗出现
2. 调用 `/api/task/workflow/issue-fix/resolve`（workflow_id 从当前路由获取）
3. 用户输入改动要求，编辑器实时展示组合结果
4. 点击"复制到剪贴板"

### 2.3 API 工具函数

**文件**: `web/src/utils/base/task_workflow.js`

新增方法 `taskWorkflowResolveIssueFix(workflowId)`，调用 `/api/task/workflow/issue-fix/resolve`。

## 三、弹窗内容拼接规则

展示内容格式：
```
[用户输入的改动要求]

[已解析的问题修改提示词模板（所有占位符已被替换）]
```

占位符替换规则与现有工作流节点一致，使用 `buildTaskWorkflowPlaceholderMap` 生成的映射。
