# dtool-butler

管家自用技能模块，提供任务管理、状态查询等工具脚本。

## 脚本

| 脚本 | 功能 | 用法 |
|------|------|------|
| `query_home_tasks.py` | 查询首页任务清单 | `python query_home_tasks.py --token TOKEN` |
| `query_self_testing_tasks.py` | 查询状态为「自测中」的任务 | `python query_self_testing_tasks.py --token TOKEN` |
| `list_tasks_in_status.py` | 按指定状态过滤任务（通用版） | `python list_tasks_in_status.py --token TOKEN --status "状态名"` |
| `check_home_task_fields.py` | 检查任务字段结构 | `python check_home_task_fields.py --token TOKEN` |

## 说明

- 所有脚本仅依赖 Python 标准库，无需 pip 安装
- 通过 dtool HTTP API 获取数据，默认地址 `http://localhost:17170`
- Token 参数必填
