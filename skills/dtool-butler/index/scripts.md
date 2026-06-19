# 脚本工具索引

共 9 个脚本工具。

## [dtool-api] Use when working with the dtool API module to query collections, folders, and APIs, create or update APIs, import APIs, run APIs, manage environments, or adjust API structure.

- 脚本: skills/dtool-api/scripts/sync_api_by_uri.py
- 集合/文件夹/接口的查询、创建、更新、删除、移动
- 批量导入接口定义
- 环境与变量管理
- 接口运行、结果查看、字段提取、代码生成
- 基于分支 diff 定位接口变更

## [dtool-common] Use when working with the dtool common module for shared API helpers, task utilities, code editing, or developer helper scripts.

- 脚本: skills/dtool-common/scripts/api_common.py, skills/dtool-common/scripts/code_edit.py
- 统一 dtool API 调用封装（api_common.py）
- 任务 SessionId 追加
- 通用代码编辑（精确文本替换/插入）
- 其他 dtool-* skill 的共享基础模块

## [dtool-db] Use when working with dtool database operations: list tables, query table structure, execute SQL queries or write operations (MySQL / Pgsql).

- 脚本: skills/dtool-db/scripts/db_api.py
- 查询数据库配置对应的所有表（MySQL / Pgsql）
- 查询指定表结构信息
- 执行 SELECT 查询
- 执行写入操作

## [dtool-docker] Use when working with dtool Docker operations: restart Docker Compose services or query service logs.

- 脚本: skills/dtool-docker/scripts/docker_api.py
- 重启 Docker Compose 指定服务
- 查询 Docker Compose 服务日志

## [dtool-git] Use when working with dtool Git operations: upload files, query/switch branches, pull code, or view branch diffs.

- 脚本: skills/dtool-git/scripts/git_api.py, skills/dtool-git/scripts/show_backend_branch_diff.py, skills/dtool-git/scripts/show_branch_diff.py, skills/dtool-git/scripts/show_file_changes.py, skills/dtool-git/scripts/show_file_diff.py, skills/dtool-git/scripts/show_frontend_branch_diff.py
- 上传本地文件到远程项目目录
- 查询当前分支、拉取代码、切换分支
- 查看当前分支相对基线的改动文件列表
- 查看单文件 diff
- 查看前端/后端常见文件类型的完整改动及 diff

## [dtool-know] Use when working with dtool knowledge fragment (memory) operations: update fragment content by fragment ID.

- 脚本: skills/dtool-know/scripts/memory_api.py
- 按片段ID更新知识片段内容（不修改标题）

## [dtool-notify] Use when the task involves sending DingTalk group notifications through a custom robot webhook.

- 脚本: skills/dtool-notify/scripts/send_dingtalk.py
- 发送钉钉群文本通知
- 支持普通 Webhook 模式
- 支持带签名密钥的安全模式

## [dtool-playwright] Use when working with the dtool smart-link / browser session module to open a logged-in browser session, capture request headers, take screenshots, or continue automation through MCP or Playwright.

- 脚本: skills/dtool-playwright/scripts/browser_api.py, skills/dtool-playwright/scripts/dtool_playwright_api.py, skills/dtool-playwright/scripts/screenshot_api.py
- 通过 smart-link 登录能力打开目标页面
- MCP 模式接管已登录浏览器会话
- Playwright 持久化目录模式接管浏览器
- 登录后抓取页面首个接口请求头
- 网页截图

## [dtool-workflow] Use when the task involves updating workflow node status for a workflow_id through the dtool workflow API.

- 脚本: skills/dtool-workflow/scripts/update_workflow_status.py
- 更新工作流节点状态（状态值由后端校验）
- 支持自定义步骤（custom_{id}）



## [dtool-butler] 自进化生成：任务清单里面有哪些任务正在自测中的

- 脚本: skills/dtool-butler/scripts/list_tasks_in_status.py
- 来源: 自进化生成

