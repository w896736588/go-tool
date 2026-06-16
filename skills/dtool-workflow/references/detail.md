# dtool-workflow 详细说明

## 必要约束

- 调用前，先向用户确认 `workflow_id`，以及 `base_url`、`Token`
- 必须使用合法的状态值，step key 由工作流模板动态定义（支持自定义步骤 custom_{id}），后端会校验合法性
- 只更新当前明确执行的步骤，不猜测用户未要求变更的节点
- 需要具体 step key 或脚本入口时，再去看 `scripts/update_workflow_status.py`

## 文件索引

- 工作流节点状态更新：`scripts/update_workflow_status.py`
