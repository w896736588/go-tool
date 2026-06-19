# 管家能力清单

dtool-butler 管家的核心能力如下：

## 内置命令

| 命令 | 说明 |
|------|------|
| /clean | 清除当前会话历史 |
| /init | 初始化索引文档（scripts.md、capabilities.md、apis.md） |
| /status | 查询管家与当前会话状态 |
| /help | 显示帮助信息 |

## Function Calling 工具

| 工具名 | 说明 |
|--------|------|
| file_read | 读取文件内容 |
| file_write | 创建或覆盖写入文件（自动创建父目录） |
| file_modify | 查找并替换文件中的指定文本 |
| file_delete | 删除文件 |
| http_call | 调用 dtool 的 HTTP API 接口（自动拼接基地址） |

## 索引与自进化

- scripts.md：扫描 skills/ 目录，生成脚本工具索引
- capabilities.md：管家总能力清单（本文件）
- apis.md：dtool HTTP 接口索引
- 检索：任务执行前 AI 判断索引中是否有可复用脚本
- 自进化：新脚本创建后自动追加到 scripts.md

## 会话管理

- 激活态：收到消息后激活，定时器重置
- 休眠巡检：超时无消息自动休眠并通知
- 历史管理：对话存库、新话题自动清历史、溢出提示

## 任务路由

- 简单任务（文件操作）→ Function Calling 工具循环
- 复杂任务（开发/重构）→ Agent CLI（Claude Code CLI / Codex CLI）

## 意图分析

- 模糊问题 → 自动追问 2-3 个澄清提问
- 明确意图 → 进入任务执行
- 新话题检测 → 自动清除旧历史
