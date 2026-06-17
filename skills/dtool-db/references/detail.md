# dtool-db 详细说明

## 必要约束

- 调用前，先向用户确认所需参数：`base_url`、`Token`、`mysql_id`
- 需要调用 dtool 接口时，优先使用 `Python` 脚本，不直接拼 bash 请求
- 数据库查询优先使用只读方式；涉及写入时必须确认影响范围

## 文件索引

- 数据库接口：`scripts/db_api.py`
