---
name: dtool-workflow
description: Use when the task involves updating workflow node statuses (需求分析/开发执行/接口生成/自动化测试+修复/代码检查/浏览器测试) via API — marking steps as pending/running/completed for a workflow_id.
---

# dtool 工作流程任务状态更新技能

- 提供更新工作流程节点状态的能力，一行调用即可。
- 用户会提供 **工作流程 ID** 和**当前步骤**，状态由你自行判断：
  - **开始执行某步骤时**传 `running`
  - **完成某步骤时**传 `completed`

## 步骤 key 对照表

| 步骤 key | 步骤名称 |
|---|---|
| `requirement` | 需求分析 |
| `design` | 开发执行 |
| `api-dev` | 接口生成 |
| `api-test-fix` | 自动化测试+修复 |
| `code-review` | 代码检查 |
| `browser-test` | 浏览器测试 |

## 调用方式

```python
import sys
sys.path.insert(0, r'C:\work\self\cache_manager_api\skills\dtool-workflow\scripts')
from update_workflow_status import update_workflow_status

# 开始执行步骤
update_workflow_status(base_url='http://localhost:17170', token='temptoken', workflow_id=69, step='requirement', status='running')

# 完成步骤
update_workflow_status(base_url='http://localhost:17170', token='temptoken', workflow_id=69, step='requirement', status='completed')
```

## 接口说明

- **路径**: `/api/task/workflow/node-status/update`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `workflow_id` | int | 是 | 工作流程 ID |
| `step` | string | 是 | 步骤 key |
| `status` | string | 是 | 状态值（`pending`/`running`/`completed`） |

## 注意事项

调用前向用户确认 `base_url` 和 `token`（默认 `http://localhost:17170`、`temptoken`）。
