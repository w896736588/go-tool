---
name: workflow-template-system
overview: 将硬编码的工作流提示词模板改造为可配置的模板系统：新增模板表和步骤表，支持模板CRUD和步骤拖拽排序，任务关联模板，老数据自动迁移，工作流页面动态渲染。
todos:
  - id: db-migrations
    content: 创建5个SQL迁移文件：新增表(tbl_workflow_template + tbl_workflow_template_step)、修改表(tbl_home_task + tbl_task_workflow)、老数据迁移脚本
    status: completed
  - id: backend-struct
    content: 新建 struct/workflow_template.go 定义模板和步骤请求/响应结构体；修改 struct/task_workflow.go 中 PromptsSaveRequest 适配 step_prompts 格式
    status: completed
  - id: backend-common
    content: 新建 common/workflow_template.go 实现模板和步骤的 CRUD 方法、排序、老数据迁移逻辑；修改 common/task_workflow.go 新增模板步骤查询方法，修改提示词保存/更新逻辑同时支持新旧字段
    status: completed
    dependencies:
      - db-migrations
      - backend-struct
  - id: backend-controller
    content: 新建 controller/workflow_template.go 实现模板管理API handler；修改 controller/task_workflow.go 的 buildTaskWorkflowResponse 集成模板步骤、PromptsSave/Restore 适配新逻辑；修改 controller/set.go 移除旧的全局提示词配置保存
    status: completed
    dependencies:
      - backend-common
  - id: backend-router-define
    content: 修改 router.go 新增 workflowTemplate() 路由注册函数；修改 define/home_task_config.go 标记旧 HomeTaskConfigPrompt* 常量为 deprecated
    status: completed
    dependencies:
      - backend-controller
  - id: frontend-api
    content: 新建 utils/base/workflow_template.js 封装模板管理API；修改 utils/base/task_workflow.js 适配 step_prompts 格式
    status: completed
  - id: frontend-template-manager
    content: 新建 components/set/WorkflowTemplateManager.vue 模板管理组件（左侧模板列表+右侧步骤CRUD+vuedraggable拖拽排序+固定步骤锁定标识）
    status: completed
    dependencies:
      - frontend-api
  - id: frontend-pages
    content: 使用 [subagent:code-explorer] 深度探索 TaskWorkflow.vue 步骤渲染逻辑后，修改 home_task_report.vue 集成 WorkflowTemplateManager、修改 HomeTask.vue 增加模板选择、修改 TaskWorkflow.vue 实现动态节点渲染
    status: completed
    dependencies:
      - frontend-template-manager
---

## 产品概述

将当前硬编码的工作流程提示词模板改造为可配置的工作流程模板系统。支持自定义多个模板，每个模板包含可排序的步骤节点，新建任务时选择模板。

## 核心功能

1. **工作流程模板管理**：设置页面支持创建、编辑、删除多个工作流程模板，每个模板含名称、描述、子步骤列表
2. **子步骤CRUD与排序**：模板下可新增/编辑/删除子步骤，每个步骤有名称和提示词MD内容；支持vuedraggable拖拽排序
3. **固定步骤**：任务配置（task-config）、抓取需求（requirement-fetch）、问题修改（issue_fix）为固定步骤，不可删除/排序
4. **任务关联模板**：新建任务时开启工作流必须选择模板；编辑时可修改模板选择
5. **老数据迁移**：现有全局提示词配置和已有工作流实例自动迁移为一个"默认模板"，所有老任务默认使用该模板
6. **工作流页面动态渲染**：根据任务关联的模板动态渲染步骤节点栏和内容区域
7. **提示词占位符**：所有步骤的提示词模板继续支持现有占位符体系

## 技术栈

- 后端：Go + SQLite（复用现有 gsdb.GsSqlite ORM 和 gsgin 框架）
- 前端：Vue3 + Element Plus + md-editor-v3
- 拖拽排序：vuedraggable（需新增 npm 依赖）
- 数据库迁移：复用现有 `internal/app/dtool/database/YYYY/MM/` SQL 文件机制

## 实现方案

### 核心设计思路

将当前"全局提示词模板配置 + 硬编码8节点工作流"两层体系，改造为"模板表 + 步骤表 + 工作流实例JSON"三层体系。工作流实例不再存储8个固定的 `prompt_xxx` 字段，而是通过关联模板获取步骤定义，通过 `step_prompts` JSON字段存储每个步骤的提示词实例值。

### 数据库设计

#### 新增表：tbl_workflow_template

```sql
CREATE TABLE "tbl_workflow_template" (
    "id"          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"        TEXT NOT NULL DEFAULT '',
    "description" TEXT NOT NULL DEFAULT '',
    "is_default"  INTEGER NOT NULL DEFAULT 0,
    "sort_order"  INTEGER NOT NULL DEFAULT 0,
    "create_time" INTEGER NOT NULL DEFAULT 0,
    "update_time" INTEGER NOT NULL DEFAULT 0
);
```

#### 新增表：tbl_workflow_template_step

```sql
CREATE TABLE "tbl_workflow_template_step" (
    "id"             INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "template_id"    INTEGER NOT NULL DEFAULT 0,
    "name"           TEXT NOT NULL DEFAULT '',
    "step_key"       TEXT NOT NULL DEFAULT '',
    "prompt_content" TEXT NOT NULL DEFAULT '',
    "sort_order"     INTEGER NOT NULL DEFAULT 0,
    "is_fixed"       INTEGER NOT NULL DEFAULT 0,
    "create_time"    INTEGER NOT NULL DEFAULT 0,
    "update_time"    INTEGER NOT NULL DEFAULT 0
);
```

#### 修改表：tbl_home_task

新增字段：`workflow_template_id INTEGER NOT NULL DEFAULT 0`

#### 修改表：tbl_task_workflow

新增字段：`step_prompts TEXT NOT NULL DEFAULT ''`（JSON格式存储步骤提示词实例值）
格式示例：`{"plain_text_requirement":"...","requirement":"...","design":"...","custom_1":"...","issue_fix":"..."}`

### 老数据迁移策略

1. 创建"默认模板"记录（is_default=1，name="默认模板"）
2. 将当前全局配置中的8个提示词创建为模板步骤（与现有 WORKFLOW_NODES 顺序一致），固定步骤标记 is_fixed=1
3. 将所有已有任务的 `workflow_template_id` 设为默认模板ID
4. 将所有已有工作流实例的8个 `prompt_xxx` 字段值迁移到 `step_prompts` JSON中
5. 迁移完成后，旧的 `prompt_xxx` 字段保留不删除（向后兼容），写入时同时写新字段和旧字段

### API设计

#### 新增模板管理API

- `POST /api/workflow/template/list` — 获取所有模板列表（含步骤）
- `POST /api/workflow/template/save` — 创建/更新模板
- `POST /api/workflow/template/delete` — 删除模板
- `POST /api/workflow/template/step/save` — 创建/更新步骤
- `POST /api/workflow/template/step/delete` — 删除步骤
- `POST /api/workflow/template/step/sort` — 步骤排序

#### 修改现有API

- `TaskWorkflowCreateOrGet` — 响应中增加 template 和 steps 信息
- `TaskWorkflowPromptsSave` — 改为按 step_key 保存 step_prompts JSON
- `TaskWorkflowPromptsRestore` — 从模板步骤的 prompt_content 还原
- `HomeTaskSave` — 增加 workflow_template_id 参数

### 前端改造要点

- **设置页面**：左侧模板列表 + 右侧步骤编辑区（vuedraggable 拖拽排序）
- **任务表单**：工作流开关下方新增模板选择下拉框
- **工作流页面**：节点栏从 `workflowNodes`（模板步骤动态生成）渲染，固定步骤保持原有渲染逻辑，自定义步骤统一为：提示词编辑器+执行按钮+执行历史按钮
- **兼容**：`prompt_type` 保持不变，自定义步骤使用 `custom_{step_id}` 格式

### 性能与兼容性

- 模板步骤数量有限（不超过20个），一次性加载，无需分页
- 旧 `prompt_xxx` 字段保留不删除，读取时优先读新字段 `step_prompts`，为空则回退读旧字段
- 写入时同时写新旧字段，确保任意时刻回滚到旧版本仍可正常读取

## 目录结构

```
internal/app/dtool/
├── common/
│   ├── task_workflow.go              # [MODIFY] 新增模板步骤读取方法、修改提示词保存逻辑支持 step_prompts
│   └── workflow_template.go          # [NEW] 模板和步骤的数据库操作方法（CRUD+排序+迁移）
├── controller/
│   ├── task_workflow.go              # [MODIFY] 修改 buildTaskWorkflowResponse、PromptsSave/Restore 适配模板
│   ├── workflow_template.go          # [NEW] 模板管理API handler（list/save/delete/step操作）
│   └── set.go                        # [MODIFY] HomeTaskConfigSave 移除旧的全局提示词配置项
├── struct/
│   ├── task_workflow.go              # [MODIFY] PromptsSaveRequest 改为 step_prompts 格式
│   └── workflow_template.go          # [NEW] 模板和步骤请求/响应结构体
├── define/
│   └── home_task_config.go           # [MODIFY] 标记8个 HomeTaskConfigPrompt* 常量为 deprecated
├── database/2026/06/
│   ├── 20260614160000-workflow_template.sql
│   ├── 20260614160100-workflow_template_step.sql
│   ├── 20260614160200-home_task_workflow_template_id.sql
│   ├── 20260614160300-task_workflow_step_prompts.sql
│   └── 20260614170000-workflow_template_migration.sql
├── router.go                         # [MODIFY] 新增 workflowTemplate() 路由注册函数

web/src/
├── components/
│   ├── TaskWorkflow.vue              # [MODIFY] 节点栏动态渲染（workflowNodes 从模板步骤生成）
│   ├── HomeTask.vue                  # [MODIFY] 任务表单增加模板选择下拉框
│   └── set/
│       ├── home_task_report.vue      # [MODIFY] "工作流提示词模板"tab 改造为模板管理界面（嵌入 WorkflowTemplateManager）
│       └── WorkflowTemplateManager.vue  # [NEW] 模板管理组件（左侧模板列表+右侧步骤CRUD+拖拽排序）
├── utils/base/
│   ├── task_workflow.js              # [MODIFY] 修改 PromptsSave 为 step_prompts 格式、PromptsRestore 适配
│   └── workflow_template.js          # [NEW] 模板管理API封装
```

## Agent Extensions

### SubAgent

- **code-explorer**
- 用途：在实现过程中进一步探索 TaskWorkflow.vue 中8个步骤节点的完整渲染分支（template 中的 v-if 条件链）和事件处理逻辑，确保动态渲染方案覆盖所有边界情况
- 预期产出：确认各步骤节点的交互差异（如 requirement-fetch 有抓取触发逻辑、issue_fix 有弹窗交互等），为动态渲染提供精确参照