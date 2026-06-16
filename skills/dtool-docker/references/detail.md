# dtool-docker 详细说明

## 必要约束

- 调用前，先向用户确认所需参数：`base_url`、`Token`、`docker_id`、`service`
- 需要调用 dtool 接口时，优先使用 `Python` 脚本，不直接拼 bash 请求
- 重启 Docker 服务前，先让用户明确 `docker_id` 和 `service`
- 查看日志时禁止使用 `-f` / `--follow` 参数

## 文件索引

- Docker 接口：`scripts/docker_api.py`
