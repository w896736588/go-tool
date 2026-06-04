---
name: dtool-common
description: Use when working with the dtool common module for remote file upload, Git operations, database queries, Docker log lookup, knowledge fragment updates, or shared helper scripts.
---

# dtool-common

## 这个 skill 可以做什么

- 上传本地文件到远程项目目录
- 查询 Git 当前分支、拉取代码、切换分支
- 查询数据库表列表、表结构、执行 SQL 查询或有限写入
- 查询 Docker Compose 服务日志
- 按文件路径更新知识片段
- 登录后抓取页面首个接口请求头
- 查看当前分支相对基线分支的改动文件和单文件 diff
- 提供通用代码编辑脚本和共享 API 调用脚本

## 必要约束

- 调用 dtool 前，先向用户确认所需参数：`base_url`、`Token`，以及任务相关的 `git_id`、`mysql_id`、`docker_id`
- 需要调用 dtool 接口时，优先使用 `Python` 脚本，不直接拼 bash 请求
- 数据库查询优先使用只读方式；涉及写入时必须确认影响范围
- 使用 Git 相关能力时，不假设默认分支名，由用户明确指定
- 需要具体参数、接口路径或脚本用法时，再去看 `scripts/` 下文件

## 细节位置

- 通用 dtool API 封装与示例：`scripts/dtool_common_api.py`
- 查看分支改动文件：`scripts/show_branch_diff.py`
- 查看单文件 diff：`scripts/show_file_diff.py`
- 查看前端常见文件类型的全部改动及 diff（默认排除 `dist`）：`scripts/show_frontend_branch_diff.py`
- 查看排除指定目录后的全部改动及 diff（不传排除目录时返回全仓库，默认排除 `dist`）：`scripts/show_backend_branch_diff.py`
- 文本替换型代码编辑：`scripts/code_edit.py`
