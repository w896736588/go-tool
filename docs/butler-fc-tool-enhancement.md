# 管家 FC 工具增强方案

## 目标

实现用户预期的完整智能链路：**索引匹配 → 代码库检索 → API 发现 → 接口调用 → 任务总结与自进化**

以"查看 common3 当前远程分支是什么"为例：

```
用户提问
  │
  ├─ 1. 索引匹配 → 在 scripts.md 中检索 git 相关脚本
  │     └─ 命中 dtool-git/scripts/git_api.py
  │
  ├─ 2. 读取 scripts.md / apis.md → 发现可用 API
  │     ├─ /api/GitConfigList → 获取 Git 配置列表（含 common3 信息）
  │     └─ /api/GitRemoteBranchList → 查询远程分支列表
  │
  ├─ 3. http_call("/api/GitConfigList") → 找到 common3 的 ssh_id 和 code_path
  │
  ├─ 4. http_call("/api/GitRemoteBranchList", {ssh_id, code_path}) → 获取远程分支列表
  │
  └─ 5. 任务总结 → 评估是否需要为本次操作建立新 skill/脚本索引
```

## Bug 修复（已完成）

### Bug 1: FC 模型文件路径解析错误

**根因：** `tool_exec.go` 中 `execFileRead` 直接用 `os.ReadFile(path)` 读取，FC 模型传相对路径 `dtool-git/git_api.py`，实际文件在 `skills/dtool-git/scripts/git_api.py`。

**修复：**
1. `worker/tool_exec.go` — 新增 `resolvePath()` 函数：相对路径读取失败时自动在 `skills/` 目录下查找
2. `core.go` — 检索命中提示改为完整相对路径 `skills/{skill_name}/scripts/{script_name}`
3. `core.go` — FC system prompt 添加 skills 目录结构说明
4. `index/init.go` — 索引中脚本列表改为完整路径格式

### Bug 2: 命令 `/init` / `/clean` 实际是工作的

这些命令正常执行，问题在于索引命中后路径传递错误，而非索引未生成。

---

## 新增工具

### 1. `http_call` — 调用 dtool HTTP API

**用途：** 允许 FC 模型直接调用 dtool 自身的 HTTP API，实现"发现 API → 调用 API"闭环。

**工具定义：**
```json
{
  "name": "http_call",
  "description": "调用 dtool 的 HTTP API 接口。所有接口均为 POST 方法，基地址已自动拼接。",
  "parameters": {
    "path": "API 路径，如 /api/GitConfigList",
    "body": "JSON 格式的请求体，如 {\"ssh_id\": \"1\"}"
  }
}
```

**实现逻辑：**
- 自动拼接基地址（如 `http://localhost:17170`）
- 发起 POST 请求，Content-Type: application/json
- 携带 `Token` 请求头（从全局配置获取，未配置时传空）
- 返回 JSON 响应体文本

**基地址来源：**
- 从 dtool 配置 `config.ini` 的 `[run] api_port` 读取
- 拼接为 `http://localhost:{api_port}`
- 通过 `ButlerEnv.DtoolBaseURL` 字段传递

---

## 涉及文件变更

| 文件 | 变更类型 | 说明 |
|------|---------|------|
| `worker/define.go` | 新增常量 | `ToolHttpCall = "http_call"` |
| `worker/tools.go` | 新增工具定义 | `http_call` 工具 schema |
| `worker/tool_exec.go` | 新增执行逻辑 | `execHttpCall()` + 基地址设置 |
| `core.go` | 修改 | 传递 `DtoolBaseURL`；更新 `fcSystemPromptSuffix` 添加工作流指引 |
| `define/butler.go` | 新增字段 | `ButlerEnv.DtoolBaseURL` |
| `business/butler_init.go` | 修改 | 从 `Env.ApiPorts` 构建基地址传入 |
| `index/capabilities.go` | 修改 | 更新能力清单包含新工具 |
| `index/init.go` | 已修复 | 脚本路径包含 `skills/{name}/scripts/` |

---

## FC System Prompt 工作流指引

FC 循环的 system prompt 补充以下工作流指导：

```
工作流程：
1. 当用户提出任务时，优先读取 apis.md 了解 dtool 提供的 HTTP 接口
2. 如果 apis.md 中有相关接口，使用 http_call 直接调用完成任务
3. 如果需要查询 Git 配置（如仓库列表），先调用 /api/GitConfigList 获取配置
4. 再根据配置中的 ssh_id、code_path 调用对应的操作接口
5. 任务完成后，简要总结执行结果。如果这是一个新的有价值的操作模式，
   评估是否需要创建新的 skill 脚本以便后续复用
```

---

## 自进化机制

任务完成后，FC 模型评估是否需要建立新的 skill/脚本索引：

**触发条件：**
- 任务使用了新的 API 组合
- 任务需要多步骤操作且可能重复
- 当前 scripts.md 中无对应脚本

**执行方式：**
- FC 模型在系统 prompt 指引下，评估任务价值
- 如果值得建立索引，调用 `file_write` 创建新的 `SKILL.md` 和 Python 脚本
- 调用 `index.EvolveAppend()` 将新脚本追加到 `scripts.md`

---

## 典型对话示例

```
用户：帮我查看一下 common3 当前远程分支是什么

管家 FC 执行：
  1. file_read("skills/dtool-git/scripts/git_api.py") → 了解脚本用法
  2. file_read("apis.md") → 发现 /api/GitConfigList 和 /api/GitRemoteBranchList
  3. http_call("/api/GitConfigList", "{}") → 获取 git_list，找到 common3 的 ssh_id=5
  4. http_call("/api/GitRemoteBranchList", "{\"ssh_id\":\"5\",\"code_path\":\"/var/www/common3\"}")
     → 返回 ["master", "develop", "feat_xxx", ...]
  5. 回复用户：common3 的远程分支有 master、develop、feat_xxx...

管家回复：
  common3 当前远程分支列表：
  - master
  - develop
  - feat_xxx
  
  💡 本次操作已记录，可复用的脚本路径：skills/dtool-git/scripts/git_api.py
```

## 实施状态

- [x] Bug 修复：文件路径解析
- [ ] 新增 `http_call` 工具
- [ ] 更新 `ButlerEnv` 携带 API 基地址
- [ ] 更新 FC system prompt 工作流
- [ ] 自进化逻辑
- [ ] 编译验证
