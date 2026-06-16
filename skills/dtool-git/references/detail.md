# dtool-git 详细说明

## 必要约束

- 调用前，先向用户确认所需参数：`base_url`、`Token`、`git_id`
- 需要调用 dtool 接口时，优先使用 `Python` 脚本，不直接拼 bash 请求
- 使用 Git 相关能力时，不假设默认分支名，由用户明确指定

## 文件索引

- Git 接口：`scripts/git_api.py`
- 分支改动文件：`scripts/show_branch_diff.py`
- 单文件 diff：`scripts/show_file_diff.py`
- 本地文件变更：`scripts/show_file_changes.py`
- 前端改动汇总：`scripts/show_frontend_branch_diff.py`
- 后端改动汇总：`scripts/show_backend_branch_diff.py`
