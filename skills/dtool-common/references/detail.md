# dtool-common 详细说明

## 必要约束

- 调用 dtool 前，先向用户确认所需参数：`base_url`、`Token`
- 需要调用 dtool 接口时，优先使用 `Python` 脚本，不直接拼 bash 请求
- 需要具体参数、接口路径或脚本用法时，再去看 `scripts/` 下文件

## 文件索引

- 通用 API 封装：`scripts/api_common.py`
- 代码编辑：`scripts/code_edit.py`
