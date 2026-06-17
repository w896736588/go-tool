# dtool-know 详细说明

## 必要约束

- 调用前，先向用户确认所需参数：`base_url`、`Token`、`workflow_id`
- 需要调用 dtool 接口时，优先使用 `Python` 脚本，不直接拼 bash 请求
- 更新片段时，后端会校验片段是否属于指定工作流

## 文件索引

- 知识片段更新：`scripts/memory_api.py`
