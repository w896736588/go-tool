# 管家通用 API 智能调用方案

> 目标：让 AI 管家能够理解 dtool 项目所有可用接口，根据用户问题自动匹配功能、查找或编写脚本、调用接口获取/执行数据，缺少信息时主动向用户确认。
> 创建时间：2026-06-18

---

## 一、目标概述

实现管家的完整智能链路：**索引匹配 → 功能定位 → 脚本复用/生成 → API 调用 → 执行 → 自进化**

核心设计理念：
- **全面索引**：`/init` 生成 dtool 所有接口和功能的完整索引（不只是预定义的 65 个，而是覆盖全部路由组）
- **智能匹配**：管家根据用户问题自动分析，匹配相关 API 能力域和已有脚本
- **脚本复用优先**：已有脚本（skills/ 下 + 自进化生成的）直接调用
- **按需生成**：无匹配脚本时，管家编写**通用性、可复用**的 Python 脚本封装 API 调用链
- **主动确认**：API 所需参数用户未提供时，向用户追问澄清
- **持续进化**：新脚本自动回写 `scripts.md`，后续可复用

```
用户问题
  │
  ├─ 1. 意图分析 → 提取功能域（Git / MySQL / Redis / 进程 / 配置 / ...）
  │
  ├─ 2. 索引匹配 → apis.md 定位相关接口组
  │     └─ 同时检索 scripts.md 看是否有现成脚本
  │
  ├─ 3. 脚本判定
  │     ├─ 有匹配脚本 → 直接调用脚本（脚本内部调 http_call 链式完成）
  │     └─ 无匹配脚本 → 管家编写通用 Python 脚本 → 调 http_call 执行
  │
  ├─ 4. 参数确认
  │     ├─ 用户已提供所有必要参数 → 继续执行
  │     └─ 缺少关键参数 → 向用户提问（如"您要操作哪个仓库？"）
  │
  ├─ 5. API 调用链
  │     ├─ 查询型：调接口获取数据 → 整理回复
  │     └─ 操作型：调接口执行 → 验证结果 → 回复
  │
  └─ 6. 自进化 → 新脚本回写 scripts.md（标记来源：自进化生成）
```

---

## 二、已完成工作

### 2.1 Bug 修复

| Bug | 根因 | 修复 |
|-----|------|------|
| FC 文件路径解析错误 | `tool_exec.go` 直接用 `os.ReadFile(path)`，FC 模型传相对路径 `dtool-git/git_api.py`，实际文件在 `skills/dtool-git/scripts/git_api.py` | 新增 `resolvePath()`：相对路径读失败时自动在 `skills/` 下查找；索引中脚本列表改为完整路径格式 |
| 命令 `/init` `/clean` 实际已工作 | 索引命中后路径传递错误，非索引未生成 | 路径格式修正后命令正常 |

### 2.2 `http_call` 工具（已实现）

允许 FC 模型直接调用 dtool HTTP API，实现"发现 API → 调用 API"闭环：

```
工具名: http_call
参数:
  path - API 路径，如 /api/GitConfigList
  body - JSON 请求体，如 {}、{"ssh_id":"5"}
基地址: 从 config.ini [run] api_port 自动拼接 http://localhost:{port}
```

---

## 三、接口与功能索引增强

### 3.1 现状

当前 `index/apis.go` 预定义了 65 个常用 API 接口。dtool 实际注册了 **28 个路由组**、合计 **100+ 个接口**，涵盖：

| 功能域 | 路由组 | 接口数 | 典型能力 |
|--------|--------|--------|---------|
| Git 管理 | `gitRouter` | 18 | 分支查询/切换、远程分支列表、Commit 日志、仓库配置列表、MR 创建 |
| MySQL 管理 | `mysqlRouter` | 4 | 表列表、表结构、SQL 查询/执行 |
| Redis 管理 | `redisRouter` | 12 | Key 搜索/类型/删除、String 读写、TTL 编辑、缓存创建 |
| 进程管理 | `toolRouter` | 8 | 端口进程列表、进程 Kill/启停/重启、日志尾行 |
| Supervisor | `supervisorRouter` | 8 | 进程组状态、配置查看、启停/全量重启 |
| SSH/服务器 | `sshRouter` + base | 10+ | SSH 配置 CRUD、服务器列表、Shell 执行、文件上传 |
| PostgreSQL | `pgSqlRouter` | 5 | 表列表、表结构、SQL 查询/执行 |
| Docker 管理 | `dockerRouter` | 5 | 容器/镜像列表、启停/删除 |
| MCP 配置 | `mcpRouter` | 10+ | MCP 配置 CRUD、连接测试、状态 |
| 定时任务 | `cronRouter` | 3+ | 任务列表、启停、日志 |
| 配置管理 (Set) | 各 `setXxxRouter` | 30+ | SSH/Git/AI/记忆库/定时/管家等配置 CRUD |
| 首页任务 | `homeTaskRouter` | 6+ | 任务 CRUD、状态、统计 |
| 任务工作流 | `taskWorkflowRouter` | 15+ | 工作流创建/推进、Chat 对话、计划 CRUD |
| 记忆片段 | `memoryFragmentRouter` | 12+ | 片段 CRUD、搜索、分类树、分享 |
| 智能链接 | `smartLinkRouter` | 5+ | 链接列表、编辑、Profile 管理 |
| AI 配置 | `aiRouter` | 10+ | Agent CLI、模型、服务商、知识库 CRUD |
| 用户/变量 | 分散 | 5+ | 用户列表、变量 CRUD |

### 3.2 增强方案

将 `apis.md` 从**预定义 65 个**改为**动态扫描全部路由**：

**方案 A：运行时扫描路由**（推荐）
- 在 `/init` 时调用 `http_call("/api/Set/ButlerApiIndex", ...)` 新增专用接口返回所有已注册路由及描述
- 优势：始终与代码同步，新增路由自动覆盖
- 需要：后端新增 `/api/Set/ButlerApiIndex` 返回路由元数据列表

**方案 B：编译期静态生成**
- 扩展 `index/apis.go`，根据 `router.go` 的路由注册完整映射表（手动维护）
- 优势：无需新增接口，简单直接
- 劣势：路由变更需同步更新映射表

> 建议先走方案 B（快速落地），后续迁移到方案 A。

### 3.3 APIs 索引格式增强

当前 `apis.md` 格式过于简单（只有路径+描述），增强为包含参数说明：

```markdown
## Git 管理

### /api/GitConfigList — Git 配置列表
- 入参: `{}`（无参数）
- 返回: `[{id, name, ssh_id, code_path, ...}]`
- 用途: 获取所有已配置的 Git 仓库信息

### /api/GitRemoteBranchList — 远程分支列表
- 入参: `{"ssh_id": "1", "code_path": "/var/www/project"}`
- 返回: `["master", "develop", "feat_xxx", ...]`
- 用途: 查询指定仓库的远程分支列表

### /api/GitChangeBranch — 切换分支
- 入参: `{"ssh_id": "1", "code_path": "/var/www/project", "branch": "develop"}`
- 返回: `{success, message}`
- 用途: 切换仓库到指定分支
- ⚠️ 危险操作：修改工作目录

### /api/MysqlQuery — SQL 查询
- 入参: `{"ssh_id": "1", "database": "mydb", "sql": "SELECT * FROM users LIMIT 10"}`
- 返回: `[{columns: [...], rows: [[...], ...]}]`
- 用途: 在指定 MySQL 实例上执行查询

...
```

---

## 四、智能匹配与路由

### 4.1 匹配流程

```
用户问题 → 意图分析（Phase 3 已有）→ 提取操作意图 + 操作对象
          ↓
  操作意图分类：
    - 查询类：查看/列出/搜索/显示 → 只读 API
    - 操作类：创建/修改/删除/切换/重启 → 读写 API
    - 分析类：统计/对比/汇总 → 多个 API 组合
          ↓
  操作对象匹配：
    - Git 相关 → 匹配 gitRouter 下的接口
    - 数据库相关 → 匹配 mysqlRouter / pgSqlRouter / redisRouter
    - 进程相关 → 匹配 toolRouter / supervisorRouter
    - 配置相关 → 匹配 Set 类接口
    - ...
          ↓
  → 将匹配结果（相关 API 路径 + 参数 schema）注入 FC system prompt
```

### 4.2 多步骤链路示例

```
用户：帮我把 common3 远程的 develop 分支最新 5 条 commit 列出来

管家分析：
  对象=Git(common3) → 操作=列出commit
  ↓
1. http_call("/api/GitConfigList", {}) → 找到 common3 的 ssh_id=5, code_path=...
2. 参数不足 → 回问用户："common3 有多个仓库配置（ID: 2,5），请问是哪个？"
3. 用户确认 → 继续
4. http_call("/api/GitRemoteBranchList", {ssh_id, code_path}) → 确认 develop 存在
5. http_call("/api/GitCommitLog", {ssh_id, code_path, branch:"develop", limit:"5"})
6. 整理回复
```

---

## 五、脚本生成与自进化

### 5.1 何时生成新脚本

管家在以下情况应编写独立的 Python 脚本（而非零散的 `http_call` 调用）：

| 场景 | 说明 | 示例 |
|------|------|------|
| 多步骤 API 调用链 | 涉及 2 个以上 API 的串联调用 | 查询仓库 → 查分支 → 查 commit → 汇总 |
| 数据需要处理/转换 | API 返回原始数据需要二次加工 | Git 分支列表按最后活跃时间排序 |
| 可预见的复用场景 | 操作模式通用，其他仓库/实例可复用 | 任意仓库"查分支 + commit" |
| 需要错误处理/重试 | API 调用可能失败需要容错 | SSH 连接超时重试 |

> **不在以下情况生成脚本**：单次 `http_call` 即可完成的简单查询（如"列出 Redis key"）。

### 5.2 脚本规范

所有自进化生成的脚本必须：

1. **通用性**：参数化（不硬编码仓库名、SSH ID、数据库名），通过命令行参数或环境变量传入
2. **独立可运行**：`python xxx.py --arg1 val1 --arg2 val2` 即可执行
3. **标准输出**：打印结果 JSON 到 stdout，错误到 stderr
4. **存放位置**：`skills/dtool-butler/scripts/`（管家专用脚本目录）
5. **命名规范**：`{功能域}_{操作}.py`，如 `git_list_commits.py`、`redis_search_key.py`
6. **附带 SKILL.md**：描述功能、参数、示例

#### 脚本模板示例

```python
#!/usr/bin/env python3
"""
git_list_commits.py — 查询指定 Git 仓库的 commit 日志
用途: 通用 Git commit 日志查询，可指定分支和条数
"""
import json, sys, argparse, requests

BASE_URL = "http://localhost:17170"
TOKEN = "temptoken"
HEADERS = {"Content-Type": "application/json", "Token": TOKEN}

def http_call(path, body=None):
    resp = requests.post(f"{BASE_URL}{path}", json=body or {}, headers=HEADERS)
    resp.raise_for_status()
    return resp.json()

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo_name", required=True, help="仓库名称（如 common3）")
    parser.add_argument("--branch", default="HEAD", help="分支名")
    parser.add_argument("--limit", type=int, default=10, help="返回条数")
    args = parser.parse_args()

    # 1. 查找仓库配置
    config_list = http_call("/api/GitConfigList")
    repo = next((r for r in config_list if r.get("name") == args.repo_name), None)
    if not repo:
        print(json.dumps({"error": f"未找到仓库: {args.repo_name}"}), file=sys.stderr)
        sys.exit(1)

    # 2. 查询 commit 日志
    result = http_call("/api/GitCommitLog", {
        "ssh_id": str(repo["ssh_id"]),
        "code_path": repo["code_path"],
        "branch": args.branch,
        "limit": str(args.limit)
    })
    print(json.dumps(result, indent=2, ensure_ascii=False))

if __name__ == "__main__":
    main()
```

### 5.3 自进化流程

```
FC 循环执行完成
  ↓
评估是否值得固化：
  ├─ 涉及 2+ API 调用 → 值得
  ├─ 单一 API 但数据处理复杂 → 值得
  └─ 单一 API 简单查询 → 不固化
  ↓
管家调用 file_write 创建脚本 + SKILL.md → 放 skills/dtool-butler/scripts/
  ↓
调用 index.EvolveAppend() → 追加条目到 scripts.md
  ↓
后续同类问题 → 索引匹配命中 → 直接调用已生成的脚本
```

---

## 六、信息确认机制

### 6.1 触发条件

管家在以下情况**必须**向用户确认后再执行：

| 场景 | 示例 |
|------|------|
| 操作对象不明确 | "帮我切到 develop 分支" → 用户有多个仓库，需要确认是哪个 |
| 危险操作 | 删除数据库、Kill 进程、切换生产分支、执行写 SQL |
| 范围过大 | "把所有 Git 仓库都 pull 一遍" → 确认是否真的要全部 |
| 参数歧义 | 用户说"common3"但匹配到多个 SSH 配置 |

### 6.2 确认交互模式

```
管家：[分析结果] 您有 3 个 Git 仓库配置：common3-web(ID:2), common3-api(ID:5), common3-admin(ID:8)
       请问要操作哪一个？（回复编号或名称即可）

用户：common3-api

管家：[执行中...]
```

> 只读查询（如查看列表、读取数据）无需确认。非破坏性操作视情况决定。

---

## 七、FC 工具扩展

### 7.1 当前工具

| 工具 | 状态 | 说明 |
|------|------|------|
| `file_read` | ✅ | 读取文件内容 |
| `file_write` | ✅ | 写入文件 |
| `file_modify` | ✅ | 查找替换修改文件 |
| `file_delete` | ✅ | 删除文件 |
| `http_call` | ✅ | 调用 dtool HTTP API（已实现） |

### 7.2 待新增工具

| 工具 | 优先级 | 说明 |
|------|--------|------|
| `run_script` | 🔴 高 | 执行 Python 脚本并返回 stdout/stderr。让管家可以调已有脚本（skills/ 下）和自进化生成的脚本 |
| `ask_user` | 🔴 高 | 向用户发起确认问题并等待回复。用于信息不足或危险操作前的确认 |
| `web_search` | 🟡 中 | 搜索互联网获取外部信息 |
| `db_query` | 🟢 低 | 直接查询 SQLite 数据库（替代 http_call 走内存查询，更快） |

#### `run_script` 设计

```json
{
  "name": "run_script",
  "description": "执行本地 Python 脚本，返回 stdout 和 stderr 输出。脚本路径基于 skills/ 目录。",
  "parameters": {
    "path": "脚本路径，如 skills/dtool-git/scripts/git_api.py",
    "args": "命令行参数列表，如 [\"--repo_name\", \"common3\", \"--branch\", \"develop\"]",
    "timeout": "超时秒数，默认 60"
  }
}
```

#### `ask_user` 设计

```json
{
  "name": "ask_user",
  "description": "向用户提问确认，暂停当前工具循环等待用户回复。仅当缺少必要信息或需要确认危险操作时使用。",
  "parameters": {
    "question": "向用户提问的内容",
    "options": "可选选项列表，如 [\"common3-web\", \"common3-api\"]，为空则自由回答",
    "reason": "需要确认的原因（如：操作对象不明确、危险操作确认）"
  }
}
```

> `ask_user` 工具调用后，FC 循环暂停，管家将问题发送给用户，用户的回复作为 `tool_call_result` 注入回 FC 循环继续执行。

---

## 八、FC System Prompt 工作流指引更新

FC 循环的 system prompt 应包含以下完整工作流指引：

```
## 可用工具
你拥有以下工具，优先使用已有脚本和 API 完成用户任务：

1. file_read / file_write / file_modify / file_delete — 文件操作
2. http_call — 调用 dtool HTTP API（所有接口均为 POST，基地址已自动拼接）
3. run_script — 执行 Python 脚本（skills/ 目录下）

## 工作流程
1. 首先理解用户意图，确定操作对象（Git 仓库？MySQL 数据库？Redis？进程？配置？）
2. 读取 apis.md 了解 dtool 提供的相关 HTTP 接口及参数要求
3. 检查 scripts.md 是否有匹配的现成脚本
4. 如果有匹配脚本 → 用 run_script 调用，传入用户指定的参数
5. 如果无匹配脚本 → 用 http_call 逐步调用 API 完成任务
6. 如果 API 需要参数但用户未提供 → 使用 ask_user 工具向用户确认
   （只读查询不需要确认，危险操作/对象不明确时才确认）
7. 任务完成后，评估是否需要建立新的通用脚本：
   - 涉及 2 个以上 API 串联调用 → 值得固化
   - 单一 API 但数据需要复杂处理 → 值得固化
   - 单一 API 简单查询 → 不固化
8. 需要固化时，用 file_write 创建 Python 脚本到 skills/dtool-butler/scripts/ 目录
   脚本要求：通用参数化、独立可运行、打印 JSON 结果

## 重要原则
- 优先使用已有脚本，避免重复造轮子
- 新写脚本必须通用化，参数通过命令行传入，不硬编码具体值
- 危险操作（删除、Kill、切换生产分支、写 SQL）必须先 ask_user
- 任务完成后简要总结执行结果
```

---

## 九、实施计划

### Phase A：apis.md 全量索引 + 参数 Schema（1-2 天）

**目标**：`/init` 生成的 `apis.md` 覆盖 dtool 所有路由组，包含接口路径、参数说明、返回值结构

| 子任务 | 说明 |
|--------|------|
| A1 扩展 `index/apis.go` | 补全 100+ 接口定义，含必要参数和返回值描述 |
| A2 API 分组标注 | 每个接口标记所属功能域（Git / MySQL / Redis / ...） |
| A3 危险操作标注 | 标记需要用户确认的接口（删除/切换/执行/Kill 等） |
| A4 `/init` 回复更新 | 告知用户生成了多少接口索引 |

**验证**：`/init` 后 `apis.md` 包含所有路由组接口，含参数说明和危险标记

### Phase B：`run_script` + `ask_user` 工具实现（1-2 天）

**目标**：FC 工具集支持脚本执行和用户交互

| 子任务 | 说明 |
|--------|------|
| B1 `worker/define.go` | 新增 `ToolRunScript`、`ToolAskUser` 常量 |
| B2 `worker/tools.go` | 新增两个工具的 OpenAI 格式定义 |
| B3 `worker/tool_exec.go` | 实现 `execRunScript()`：调用 `os/exec` 执行 Python，捕获 stdout/stderr。实现 `execAskUser()`：暂停循环，问题发送给用户，等待回复注入 |
| B4 `worker/fc_loop.go` | 支持 `ask_user` 导致的暂停/恢复机制 |
| B5 `butler/core.go` | 处理 `ask_user` 消息：暂停当前 FC 循环 → 发送问题到钉钉 → 收到回复后恢复 FC 循环 |

**验证**：用户"查看 common3 分支"→ 多仓库时 ask_user 确认 → 收到回复后继续执行

### Phase C：自进化脚本生成（1 天）

**目标**：FC 循环完成后自动评估并生成可复用的 Python 脚本

| 子任务 | 说明 |
|--------|------|
| C1 FC system prompt 更新 | 追加脚本固化评估指引和模板 |
| C2 评估逻辑 | FC 模型判断是否值得固化（多 API 链 / 复杂处理） |
| C3 脚本生成与回写 | `file_write` 创建脚本到 `skills/dtool-butler/scripts/`，`evolve.go` 追加到 `scripts.md` |
| C4 `run_script` 路径优先级 | 优先查 `skills/dtool-butler/scripts/`（自进化），再查 `skills/dtool-*/scripts/`（内置） |

**验证**：用户完成一个涉及 2+ API 调用的任务 → 自动生成脚本 → 下次同类问题直接命中脚本

### Phase D：信息确认交互完善（1 天）

**目标**：管家在适当时机主动向用户确认，不会盲目执行

| 子任务 | 说明 |
|--------|------|
| D1 危险操作标记传播 | `apis.md` 中的危险标记 → FC system prompt 引导→ 自动触发 ask_user |
| D2 参数不足检测 | 当 FC 判断 API 需要的参数用户未提供 → 调用 ask_user |
| D3 确认超时处理 | 用户长时间未回复 → 取消任务并通知 |

**验证**：危险操作前暂停确认，用户确认后执行；取消后正确回收

---

## 十、典型对话示例

### 示例 1：查询类 — 已有脚本

```
用户：帮我看看 common3 的远程分支

管家：
  意图分析: 操作对象=Git(common3), 操作=查询远程分支
  脚本检索: scripts.md 命中 skills/dtool-git/scripts/git_api.py
  → run_script("skills/dtool-git/scripts/git_api.py", ["--repo_name", "common3", "--cmd", "remote_branch"])
  → 脚本内部: http_call(GitConfigList) → http_call(GitRemoteBranchList)
  → 回复: common3 远程分支有 master、develop、feat_xxx、feat_yyy
```

### 示例 2：操作类 — 需要确认

```
用户：帮我把 common3 切到 develop 分支

管家：
  意图分析: 操作对象=Git(common3), 操作=切换分支 ⚠️ 危险操作
  → http_call("/api/GitConfigList") → 找到 3 个 common3 配置
  → ask_user("您有 3 个 common3 仓库：common3-web(ID:2)、common3-api(ID:5)、common3-admin(ID:8)，要切换哪个？")
  
用户回复：api

管家：
  → http_call("/api/GitRemoteBranchList", {ssh_id:"5", ...}) → 确认 develop 存在
  → ask_user("确认将 common3-api 切换到 develop 分支？当前可能在 feat 分支上。")
  
用户回复：确认

管家：
  → http_call("/api/GitChangeBranch", {ssh_id:"5", code_path:"...", branch:"develop"})
  → 回复: ✅ common3-api 已切换到 develop 分支
  
  💡 本次操作已生成通用脚本: skills/dtool-butler/scripts/git_switch_branch.py
    下次可直接使用: python git_switch_branch.py --repo_name xxx --branch yyy
```

### 示例 3：无脚本 — 自进化生成

```
用户：帮我看看所有 Git 仓库当前分别在哪条分支上

管家：
  意图分析: 操作对象=Git(全部仓库), 操作=查询当前分支
  脚本检索: scripts.md 无匹配脚本（现有脚本只支持单仓库查询）
  ↓
  自写脚本执行:
    1. http_call("/api/GitConfigList") → 获取所有仓库列表
    2. 遍历每个仓库 → http_call("/api/GitCurrentBranch", {id})  → 聚合结果
    3. 整理为表格回复
  
  回复：
    common3-web(ID:2): master
    common3-api(ID:5): feat_login_20250615
    common3-admin(ID:8): master
    ...

  💡 已生成通用脚本: skills/dtool-butler/scripts/git_all_branches.py
    下次可直接: python git_all_branches.py 查看所有仓库分支状态
```

---

## 十一、实施状态

| Phase | 模块 | 状态 | 产出 |
|-------|------|------|------|
| - | Bug 修复（文件路径解析） | ✅ | `resolvePath()`, 路径格式修正 |
| - | `http_call` 工具 | ✅ | 工具定义 + 执行逻辑 |
| A | apis.md 全量索引 + 参数 Schema | ⬜ 待开发 | 100+ 接口完整索引含参数描述 |
| B | `run_script` + `ask_user` 工具 | ⬜ 待开发 | FC 工具集支持脚本执行和用户交互 |
| C | 自进化脚本生成 | ⬜ 待开发 | FC 完成后自动评估并固化脚本 |
| D | 信息确认交互完善 | ⬜ 待开发 | 危险操作确认、参数不足追问、超时处理 |
