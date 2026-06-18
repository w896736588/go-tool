# AI 管家（dtool-butler）设计与实施计划

> 基于流式机器人启动、能自进化的 dtool 智能管家。
> 创建时间：2026-06-17

---

## 一、已确认决策

| # | 决策点 | 结论 |
|---|--------|------|
| 1 | 子管家执行层 | 可选，关联 dtool 现有 `tbl_agent_cli` 配置；简单文件操作走 Function Calling（直接调脚本），复杂开发走 Agent CLI |
| 2 | 机器人平台 | 先仅支持钉钉，预留飞书/企微扩展位 |
| 3 | 消息收发 | **钉钉 Stream 模式（WebSocket 长连接）接收**，无需公网 IP；发送用 webhook 主动 POST |
| 4 | 索引文档 | 用固定 md 文件，**放记忆库目录**（`{memoryDbPath}/butler/index/`，便于 Git 自动同步） |
| 5 | 工程结构 | 管家代码合入 `internal/app/dtool/butler/`，与 dtool 同进程运行；类型定义放 `dtool/define/butler.go`；配置走 dtool 项目；管家使用独立 SQLite（butler.db） |
| 6 | migration | **管家表 SQL 放 `dtool/database_butler/`，dtool 启动时执行**（与主库 migration 隔离，记录表为 `tbl_butler_database_up`） |
| 7 | 索引文档位置 | `{memoryDbPath}/butler/index/` 下 3 个 md（capabilities.md / scripts.md / apis.md） |

---

## 二、技术现实说明（钉钉 Stream）

- 钉钉机器人接收消息有两种官方模式：**HTTP 模式**（需公网回调地址，即 webhook）和 **Stream 模式**（WebSocket 长连接，无需公网）。
- "长轮询、不用 webhook 回调"的诉求与 Stream 模式意图一致：主动维持长连接、无需公网 IP、被动接收消息。
- Go 官方 SDK：`github.com/open-dingtalk/dingtalk-stream-sdk-go` v0.9.1，已加入 go.mod。
- 关键 API：
  - `client.NewStreamClient(client.WithAppCredential(client.NewAppCredentialConfig(appKey, appSecret)))`
  - `cli.RegisterChatBotCallbackRouter(handler)`，handler 签名 `func(ctx, *chatbot.BotCallbackDataModel) ([]byte, error)`
  - `cli.Start(ctx)` 非阻塞，内部起 goroutine 维持长连接，自动重连
- 消息回复：机器人消息携带 `SessionWebhook`（临时有效），用 `chatbot.NewChatbotReplier().SimpleReplyText(ctx, sessionWebhook, content)` 直接回复，无需单独配置发送地址。
- 主动推送（打招呼/休眠通知等无 incoming 场景）：通过群机器人 webhook_url 主动 POST（加签），复用 dtool `webhook_notify.go` 思路。

---

## 三、整体架构

```text
┌──────────────────────────────────────────────────────────────┐
│  钉钉（用户在群里/单聊发消息）                                │
└──────────────────────────────────────────────────────────────┘
              ↑ Stream 长连接接收            ↓ webhook 主动 POST 发送
┌──────────────────────────────────────────────────────────────┐
│  cmd/dtool  （统一进程，管家已合入）                           │
│                                                                │
│  ┌─ butler/     管家核心（角色/激活态/命令/意图/历史）          │
│  ├─ butler/bot/     钉钉 Stream 网关（收发 + 预留多平台）      │
│  ├─ butler/worker/  子管家调度（FC 工具循环 / Agent CLI / 验收）│
│  ├─ butler/index/   索引文档（固定 md，生成/检索/自进化回写）   │
│  ├─ define/butler.go  管家类型与常量                           │
│  ├─ business/butler_init.go  管家运行时（ButlerRuntime）       │
│  ├─ controller/set_butler.go 管家配置 CRUD                    │
│  ├─ database_butler/  管家 migration SQL                      │
│  │                                                                │
│  │  ── 以下为 dtool 原有模块 ──                                   │
│  ├─ common/     公共数据访问与工具                              │
│  ├─ component/  全局单例组件                                    │
│  ├─ define/     业务类型与常量                                  │
│  ├─ controller/ HTTP 控制器                                    │
│  ├─ business/   业务逻辑层                                     │
│  └─ database/   主库 migration SQL                             │
│                                                                │
│  复用 dev_tool/internal/pkg/{p_db,p_claude,p_codex,p_common}   │
└──────────────────────────────────────────────────────────────┘
```

### 复用与边界
- **管家代码已合入 dtool 进程**：原来在 `internal/app/dtool-butler/` 下的代码已全部迁移到 `internal/app/dtool/butler/`，与 dtool 同进程运行，无需独立部署。
- **配置管理在 dtool**：dtool 新增 `controller/set_butler.go` + 前端配置页，CRUD 管家配置表。管家只读这些表。
- **migration 隔离**：管家表 SQL 放 `internal/app/dtool/database_butler/`，dtool 启动时执行，记录表为 `tbl_butler_database_up`，与 dtool 的 `tbl_database_up` 完全隔离。
- **管家与 dtool 互调**：同进程内直接调用，无需 HTTP 互调。
- **管家使用独立数据库文件**（butler.db），与主库（dtool.db）文件隔离，但共用同一目录。

---

## 四、项目目录结构

```
cmd/dtool/
  main.go                        # 统一入口（管家已合入 dtool 进程）
internal/app/dtool/
  config.go                      # dtool 初始化编排，含管家库初始化（initButlerSqlite）
  butler/
    core.go                      # 管家主循环：打招呼/消费消息/FC回复/休眠巡检
    session.go                   # 会话/激活态管理 + 30min 休眠回收
    history.go                   # 历史对话存储/查询/清理
    role.go                      # 角色系统：加载 persona/tone/system_prompt
    command.go                   # 内置命令：clean/init/status/help
    intent.go                    # 意图分析 + 自动追问
    bot/
      gateway.go                 # 统一网关接口（Gateway），预留多平台
      dingtalk_gateway.go        # 钉钉 Stream 接收 + webhook 发送
    worker/
      define.go                  # 工具名称常量
      tools.go                   # 基础文件工具定义 OpenAI 格式
      tool_exec.go               # 工具执行逻辑（Go 原生文件 I/O）
      fc_loop.go                 # Function Calling 工具循环
      verify.go                  # 监督验收
      dispatcher.go              # 任务路由：简单→FC，复杂→Agent CLI
      agent_cli.go               # 复用 p_claude/p_codex 执行复杂任务
    index/
      doc.go                     # 索引文档读写基础
      init.go                    # init 命令：扫描 skills/ 生成全部 3 个索引文件
      capabilities.go            # capabilities.md 生成（管家总能力清单）
      apis.go                    # apis.md 生成（dtool HTTP 接口索引）
      retrieve.go                # 检索匹配
      evolve.go                  # 自进化：新脚本回写索引
  define/
    butler.go                    # 管家类型/常量/状态枚举（ButlerEnv/ButlerConfigItem等）
  business/
    butler_init.go               # 管家运行时（ButlerRuntime），管理机器人连接与管家核心
  controller/
    set_butler.go                # 管家配置 CRUD 接口（BotConfig/Role/Config/Message）
  database_butler/               # 管家 migration SQL（按年月组织）
    2026/06/20260617100000_butler_init.sql
    2026/06/20260618100000_butler_conn_status_and_msg_bot.sql
```

---

## 五、数据库表设计（建在 butler.db，migration 由 dtool 启动时执行）

### 5.1 配置类（dtool Web 管理，管家只读）
```sql
-- 钉钉机器人配置（Stream 模式所需）
CREATE TABLE IF NOT EXISTS "tbl_butler_bot_config" (
  "id"           INTEGER PRIMARY KEY AUTOINCREMENT,
  "platform"     TEXT NOT NULL DEFAULT 'dingtalk',   -- 预留 feishu/wecom
  "name"         TEXT NOT NULL DEFAULT '',
  "app_key"      TEXT NOT NULL DEFAULT '',            -- 钉钉应用 AppKey
  "app_secret"   TEXT NOT NULL DEFAULT '',            -- 钉钉应用 AppSecret
  "robot_code"   TEXT NOT NULL DEFAULT '',            -- 机器人 robotCode
  "webhook_url"  TEXT NOT NULL DEFAULT '',            -- 发送用 webhook（主动推送）
  "secret"       TEXT NOT NULL DEFAULT '',            -- 发送加签
  "status"       INTEGER NOT NULL DEFAULT 1,
  "created_at"   INTEGER NOT NULL DEFAULT 0,
  "updated_at"   INTEGER NOT NULL DEFAULT 0
);

-- 管家角色
CREATE TABLE IF NOT EXISTS "tbl_butler_role" (
  "id"             INTEGER PRIMARY KEY AUTOINCREMENT,
  "name"           TEXT NOT NULL DEFAULT '',
  "persona"        TEXT NOT NULL DEFAULT '',   -- 定位（如"严谨的技术管家"）
  "tone"           TEXT NOT NULL DEFAULT '',   -- 语气（如"简洁专业"）
  "system_prompt"  TEXT NOT NULL DEFAULT '',   -- 完整 system prompt
  "init_greeting"  TEXT NOT NULL DEFAULT '',   -- 启动打招呼语
  "status"         INTEGER NOT NULL DEFAULT 1,
  "created_at"     INTEGER NOT NULL DEFAULT 0,
  "updated_at"     INTEGER NOT NULL DEFAULT 0
);

-- 管家运行参数
CREATE TABLE IF NOT EXISTS "tbl_butler_config" (
  "id"                       INTEGER PRIMARY KEY AUTOINCREMENT,
  "name"                     TEXT NOT NULL DEFAULT '',
  "role_id"                  INTEGER NOT NULL DEFAULT 0,   -- 关联角色
  "model_id"                 INTEGER NOT NULL DEFAULT 0,   -- 管家主模型（tbl_ai_model）
  "fc_model_id"              INTEGER NOT NULL DEFAULT 0,   -- Function Calling 用模型
  "agent_cli_id"             INTEGER NOT NULL DEFAULT 0,   -- 可选：复杂任务用的 AgentCli
  "bot_config_id"            INTEGER NOT NULL DEFAULT 0,   -- 关联机器人配置
  "active_timeout_minutes"   INTEGER NOT NULL DEFAULT 30,  -- 激活态超时
  "max_history"              INTEGER NOT NULL DEFAULT 100, -- 历史上限
  "auto_clean_on_new_topic"  INTEGER NOT NULL DEFAULT 1,   -- 新问题自动清历史
  "index_doc_path"           TEXT NOT NULL DEFAULT '',     -- 索引 md 目录（留空用默认）
  "auto_init_on_start"       INTEGER NOT NULL DEFAULT 1,   -- 启动自动 init 索引
  "status"                   INTEGER NOT NULL DEFAULT 1,
  "created_at"               INTEGER NOT NULL DEFAULT 0,
  "updated_at"               INTEGER NOT NULL DEFAULT 0
);
```

### 5.2 运行时类（管家读写）
```sql
-- 会话历史
CREATE TABLE IF NOT EXISTS "tbl_butler_message" (
  "id"          INTEGER PRIMARY KEY AUTOINCREMENT,
  "session_id"  TEXT NOT NULL DEFAULT '',        -- 会话标识（如钉钉 conversationId）
  "role"        TEXT NOT NULL DEFAULT '',        -- user/assistant/system
  "content"     TEXT NOT NULL DEFAULT '',
  "token_count" INTEGER NOT NULL DEFAULT 0,
  "topic"       TEXT NOT NULL DEFAULT '',        -- 当前主题（用于新问题判定）
  "created_at"  INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS "idx_butler_msg_session" ON "tbl_butler_message"("session_id", "id");

-- 管家任务记录（监督/执行/验收）
CREATE TABLE IF NOT EXISTS "tbl_butler_task" (
  "id"          INTEGER PRIMARY KEY AUTOINCREMENT,
  "session_id"  TEXT NOT NULL DEFAULT '',
  "title"       TEXT NOT NULL DEFAULT '',
  "status"      TEXT NOT NULL DEFAULT 'pending', -- pending/executing/verifying/done/failed
  "plan"        TEXT NOT NULL DEFAULT '',        -- 任务拆解
  "result"      TEXT NOT NULL DEFAULT '',        -- 验收结果
  "executor"    TEXT NOT NULL DEFAULT '',        -- fc/agent_cli
  "created_at"  INTEGER NOT NULL DEFAULT 0,
  "updated_at"  INTEGER NOT NULL DEFAULT 0
);
```

---

## 六、索引文档设计（固定 md，放记忆库目录）

位置：`{config.ini: base.memoryDbPath}/butler/index/`（如 `C:\work\self\dev_tool_db\memory\butler\index\`）。
管家启动时从 config.ini 读 `base.memoryDbPath` 定位，复用记忆库的 Git 自动同步能力做版本管理。

```
{memoryDbPath}/butler/index/
  capabilities.md     # 总能力清单（管家能做什么）
  scripts.md          # 脚本工具索引（skills/ 下脚本 + 自进化新脚本）
  apis.md             # dtool 可用 HTTP 接口索引
```

### scripts.md 结构
```markdown
# 脚本工具索引

## [db_api] 数据库查询
- 路径: skills/dtool-db/scripts/db_api.py
- 用途: 查询/验证数据库字段、索引
- 入参: table_name, conditions
- 出参: JSON 行数据
- 示例: python db_api.py --table tbl_user

## [git_diff] 获取分支改动
- 路径: skills/dtool-common/scripts/xxx.py
- 用途: ...
```

### 自进化流程
1. 任务来 → `retrieve.go` 把"任务 + scripts.md 摘要"喂 AI，判断是否有现成脚本
2. 命中 → 直接调用
3. 未命中 → 子管家编写新脚本 → 执行 → `evolve.go` 追加条目到 scripts.md

---

## 七、分阶段实施计划

### Phase 1：钉钉双向通信 + 管家骨架（最小闭环） ✅ 代码已完成
**目标**：启动管家 → 自动发打招呼 → 用户发消息 → 管家回复（先不接 AI，固定回复）→ 30min 无消息休眠
**检查点**：钉钉发消息管家能收到并回复；启动自动发打招呼；30min 无消息自动休眠并通知

| 子任务 | 状态 | 说明 |
|--------|------|------|
| P1-1 工程骨架 + 共用库连接 | ✅ | `main.go` + `config.go` + migration，已验证建表成功 |
| P1-2 钉钉 Stream 接收 | ✅ | `dingtalk_gateway.go` 接入 SDK，代码已完成待凭证联调 |
| P1-3 钉钉发送 | ✅ | webhook POST 发送，代码已完成待凭证联调 |
| P1-4 管家主循环 + 激活态 | ✅ | `core.go` + `session.go`，打招呼/消费/休眠巡检 |
| P1-5 历史消息存储 | ✅ | `history.go`，消息存 `tbl_butler_message` |

**Phase 1 关键发现（踩坑记录）**：
- `gsdb` 的 `QueryBySql(...).One()` 会自动追加 `LIMIT 1`，SQL 里不能再自带 `LIMIT`，否则产生 `... LIMIT 1 LIMIT 1` 语法错误。使用 `.One()` 时 SQL 不要写 `LIMIT`；需要自定义 limit 时用 `.All()` + SQL 自带 `LIMIT`。
- 钉钉 SDK 回复机制：消息携带 `SessionWebhook`（临时有效），直接用它回复，无需单独发送配置；仅主动推送（打招呼/休眠通知）需群机器人 webhook_url。

**Phase 1 联调前置条件**：
1. 在 dtool 配置 `tbl_butler_bot_config`（钉钉 app_key/app_secret/webhook_url/secret）
2. 在 dtool 配置 `tbl_butler_role`（角色/打招呼语）
3. 在 dtool 配置 `tbl_butler_config`（关联 role_id/bot_config_id，设置 active_timeout_minutes 等）
4. 钉钉开放平台创建企业内部应用并开通 Stream 模式机器人

### Phase 2：角色系统 + 内置命令 + 历史管理 ✅ 代码已完成
**目标**：管家有"人格"，支持 clean/init/status 命令，历史存库
- [x] `role.go` 加载角色 persona/tone，拼装 system_prompt
- [x] `command.go` 解析 clean/init/status/help
- [x] `history.go` clean 清当前 session + ToAiMessages 转换
- [x] 管家回复接 `AIChatStreamByModelWithMessages`（流式多轮，简洁约束）
- [x] `common/ai_chat.go` 新增 `AIChatStreamByModelWithMessages` 支持多轮对话
- [x] 验证：编译通过；clean 清历史，角色语气生效（需钉钉联调验证）

**Phase 2 关键设计决策**：
- `AIChatStreamByModelWithMessages` 接受 `[]map[string]string` messages 列表，支持多轮对话，与原 `AIChatStreamByModel`（仅 system+user）隔离，不破坏现有接口
- `BuildSystemPrompt` 优先使用角色的 `system_prompt` 字段（完整自定义），否则用 `persona + tone` 组合生成，兜底默认 prompt
- 命令路由优先级高于 AI：以 `/` 开头优先走命令解析，未命中命令则交由 AI 处理
- 钉钉无法逐 chunk 推送，AI 流式调用收集完整回复后一次性发送

### Phase 3：意图分析 + 自动追问 ✅ 代码已完成
**目标**：模糊问题返回 2-3 个澄清提问，明确则进入任务
- [x] `intent.go` 轻量 AI Chat 判断意图清晰度 + 新话题判定（JSON 格式输出）
- [x] 新问题 + `auto_clean_on_new_topic` → 自动清历史
- [x] 历史超 `max_history` → 回复末尾附加提示（建议 /clean）
- [x] `history.go` 补充 `GetRecentTopic` + `UpdateTopicBySession` 主题回填
- [x] 验证：编译通过；模糊问题自动追问，新主题自动清历史（需钉钉联调验证）

**Phase 3 关键设计决策**：
- 意图分析使用 `fc_model_id`（轻量模型），为 0 时回落 `model_id`；模型未配置时跳过分析，默认意图清晰
- AI 返回 JSON 格式 `{clear, topic, new_topic, questions}`，含容错提取逻辑（AI 可能包裹额外文字）
- 意图模糊 → 直接返回澄清提问，不进入 AI 主回复流程（避免二次 AI 调用浪费）
- 新话题检测后自动清历史，同时重新存当前用户消息 + 回填主题到历史记录
- 历史溢出提示采用非阻塞方式（附加在 AI 回复末尾），不中断对话流程

### Phase 4：子管家执行层（基础文件工具 + Function Calling） ✅ 代码已完成
**目标**：用户发文件操作任务 → FC 工具循环执行 → 汇报
- [x] `worker/tools.go` 注册 file_read/write/modify/delete（Go 原生文件 I/O 实现）
- [x] `worker/tool_exec.go` 工具执行逻辑（读取、写入、查找替换、删除文件）
- [x] `worker/fc_loop.go` AI Chat + tools 循环（Function Calling 循环）
- [x] `worker/verify.go` 验收产出（文件存在/内容验证/删除验证）
- [x] `common/ai_chat.go` 新增 `AIChatByModelWithTools` 方法支持 Function Calling
- [x] `butler/core.go` 集成 FC 循环，替换原 `aiReply` 为 `fcReply`，工具使用时自动创建任务记录
- [x] 验证：编译通过

**Phase 4 关键设计决策**：
- 文件工具使用 Go 原生 `os.ReadFile/os.WriteFile/os.Remove` 实现，而非调用 Python skill 脚本，减少进程间通信开销
- FC 循环使用 `fc_model_id`（Function Calling 用模型），为 0 时回落 `model_id`；与意图分析复用同一模型选择逻辑
- FC 循环使用非流式 AI 请求（需解析完整 `tool_calls` JSON），最大迭代 10 次防止无限循环
- 工具调用完成后自动创建 `tbl_butler_task` 任务记录（仅当有工具调用时），记录任务标题、结果、使用的工具列表
- `file_modify` 采用查找替换模式（`strings.Replace` 替换第一个匹配项），与 `code_edit.py` 思路一致
- FC system prompt 在角色 prompt 基础上追加工具使用指引，AI 可自行决定是否调用工具（普通对话不需要工具）
- `extractContentAndToolCalls` 从 AI 响应中同时提取 `content` 和 `tool_calls`，支持 AI 同时返回文本和工具调用的场景

### Phase 5：Agent CLI 复杂任务执行 ✅ 代码已完成
**目标**：复杂开发任务走 `tbl_butler_config.agent_cli_id` 指定的 AgentCli
- [x] `worker/agent_cli.go` 复用 `p_claude.RunClaudeStream`/`p_codex.RunCodexStream`
- [x] `worker/dispatcher.go` 路由：简单→FC，复杂→Agent CLI（AI 分类判断）
- [x] `butler/core.go` 集成 dispatcher，`fcReply` 先路由再执行
- [x] 验证：编译通过

**Phase 5 关键设计决策**：
- 任务路由使用 AI 分类判断（`fc_model_id`）：用户消息 → AI 判断 `fc`/`agent_cli` → 对应执行路径
- `agent_cli_id` 为 0 时始终走 FC（无 Agent CLI 可用），保证向后兼容
- Agent CLI 执行时将角色 system prompt 和用户消息拼合为 prompt，确保 Agent CLI 行为与管家角色一致
- Claude Code CLI：从 `settings.json` 读取模型/API 配置，使用 `RunClaudeStream` 流式执行
- Codex CLI：从 `tbl_agent_cli.config` JSON 字段读取配置（`CodexCliConfig`），使用 `RunCodexStream` 流式执行
- Agent CLI 最大执行时间 10 分钟（`agentCliTimeout`），超时自动取消
- Agent CLI 结果收集：Claude 从 `assistant`/`result` 事件提取文本，Codex 从 `item.completed`/`agent_message` 事件提取
- Agent CLI 任务记录：执行成功 → `status=done`，执行失败 → `status=failed`，`executor=agent_cli`
- 任务路由结果容错解析：`parseDispatchResult` 支持多种格式（`agent_cli`/`agent-cli`/`agentcli`，包含 `agent` 关键词），默认走 FC

### Phase 6：索引文档 + 自进化 ✅ 代码已完成
**目标**：init 生成索引，任务前检索，新脚本回写
- [x] `index/doc.go` 索引文档读写基础（路径解析、文件读写、存在性检查）
- [x] `index/init.go` 扫描 `skills/` 生成 scripts.md（解析 SKILL.md front matter + 功能索引 + 脚本列表）
- [x] `index/retrieve.go` 任务前检索匹配（AI 判断索引中是否有可复用脚本）
- [x] `index/evolve.go` 新脚本回写 scripts.md（自进化追加条目）
- [x] `butler/command.go` `/init` 命令真正触发索引生成
- [x] `butler/core.go` 启动自动初始化索引 + FC 循环前检索匹配
- [x] 验证：编译通过

**Phase 6 关键设计决策**：
- 索引路径优先使用 `config.IndexDocPath`，为空时回落到 `{memoryDbPath}/butler/index/`
- `scripts.md` 扫描生成：遍历 `skills/` 下所有子目录，解析 `SKILL.md` 的 YAML front matter（name、description）和功能索引列表，收集 `scripts/` 下的 `.py` 文件
- 启动自动初始化：`auto_init_on_start=1` 时管家启动自动生成索引，索引已存在时跳过
- `/init` 命令现在真正执行索引生成，返回 scripts.md 行数
- 任务前检索：FC 循环执行前调用 `index.Retrieve()`，用 AI 判断索引中是否有匹配脚本
- 检索命中 → 将匹配信息注入 FC 循环的 system prompt（`💡 索引匹配：找到相关脚本 xxx/yyy`）
- 自进化：`EvolveAppend()` 追加新脚本条目到 scripts.md 末尾，标记"来源：自进化生成"
- `CommandContext` 结构体新增，用于命令执行时传递 `indexPath` 和 `skillsRoot`
- Core 结构体新增 `env`、`indexPath`、`skillsRoot` 字段，`NewCore` 签名新增 `env` 参数

每个 Phase 完成后验证再推进下一阶段。

### Phase 7：dtool 端管家配置管理 ✅ 代码已完成
**目标**：dtool Web 界面 CRUD 管理管家配置表（机器人、角色、运行参数）
- [x] `controller/set_butler.go` 9 个 CRUD 接口（BotConfig/Role/Config 各 List/Add/Delete）
- [x] `router.go` 新增 `butlerRouter()` 注册 9 条 `/api/Set/Butler*` 路由
- [x] 前端 `web/src/utils/base/butler_set.js` API 模块
- [x] 前端 `web/src/components/set/butler.vue` 配置管理页（机器人/角色/管家 3 个子 Tab）
- [x] `web/src/components/Set.vue` 添加 Butler Tab
- [x] 验证：Go 编译通过 + Vue 构建通过

**Phase 7 关键设计决策**：
- BotConfig 列表返回时 app_secret 和 secret 字段自动脱敏（保留前 6 + 后 4 位，中间星号），编辑弹窗不回填密钥字段
- ButlerConfig 列表查询时关联角色、模型、Agent CLI、机器人配置名称（`role_name`、`model_name` 等虚拟字段）
- ButlerConfig 编辑弹窗的下拉数据来源：角色列表（ButlerRoleList）、AI 模型列表（AiModelList）、Agent CLI 列表（AgentCliList）、机器人配置列表（ButlerBotConfigList）
- 前端页面采用与 ai_provider.vue 相同的内部 Tab 结构，按需加载活跃 Tab 数据

### Phase 8：索引文档补充 ✅ 代码已完成
**目标**：补充 capabilities.md 和 apis.md 生成逻辑
- [x] `index/capabilities.go` GenerateCapabilitiesIndex — 管家总能力清单（内置命令、FC 工具、索引与自进化、会话管理、任务路由、意图分析）
- [x] `index/apis.go` GenerateApisIndex — dtool HTTP 接口索引（预定义 65 个接口，按功能分组展示）
- [x] `index/init.go` InitIndex 更新为同时生成 3 个索引文件（scripts.md + capabilities.md + apis.md）
- [x] `butler/command.go` /init 回复信息更新（反映 3 个文件）
- [x] 验证：Go 编译通过

**Phase 8 关键设计决策**：
- capabilities.md 为静态内容，基于管家已知能力结构生成（内置命令、FC 工具列表、会话管理、任务路由等）
- apis.md 为预定义接口列表（65 个常用 dtool API），按功能分组展示，而非运行时动态扫描（但ler与 dtool 同进程，理论上可动态扫描，但预定义方式更稳定可控）
- InitIndex 现在生成全部 3 个文件，但返回值仍为 scripts.md 内容（主要索引）

---

## 八、数据流（典型：用户发任务）

```
用户在钉钉发消息
  → DingTalkGateway Stream 接收 → 投递 IncomingMessage 到 msgChan
  → Butler Core 消费 → sessions.Activate 激活态 + 重置 30min 定时器
  → history.Append 存用户消息到 tbl_butler_message
  → (Phase 2+) 加载角色 system_prompt + 历史对话
  → (Phase 3+) intent.go 意图分析：
      ├─ 命中内置命令(clean/init) → 直接执行
      ├─ 意图不明确 → 返回澄清提问
      └─ 意图明确 → 任务拆解
  → (Phase 4/5) 任务分发到 Sub-Butler：
      ├─ 检索脚本索引 → 命中已有脚本 → 直接调用
      └─ 未命中 → 子管家编写新脚本 → 执行 → 回写索引
  → (Phase 4) verify.go 验收子管家产出
  → 结果汇总 → replier.SimpleReplyText 回复到钉钉
  → history.Append 存管家回复
  → (Phase 3) 历史超阈值 → 触发清理提示/自动清理
```
