# AI 自动化测试方案

> 目标：从任务清单的 `tapd_url` 出发，自动生成需求 MD → AI 解析需求生成测试计划 → 自动执行测试 → 验证接口是否符合需求

---

## 一、背景与目标

### 1.1 现状

当前项目（dtool）已有完整的开发工具链：

- 任务清单支持 `tapd_url`，可自动抓取 TAPD 需求页并转为 MD
- 内置 Playwright 浏览器自动化（Smart Link + Agent 模式）
- 内置接口开发模块，支持接口定义、环境管理、接口执行（`/api/ApiRun`）
- `dtool-agent` 可通过 WebSocket 远程执行 Playwright 任务

### 1.2 目标

在 AI 完成接口开发后，自动完成以下闭环：

```
tapd_url → 抓取需求MD → AI解析需求 → 生成测试计划 → 自动执行测试 → 输出测试报告
```

验证维度：
1. **接口正确性** — 接口是否按需求正常工作（状态码、返回结构、业务逻辑）
2. **需求符合度** — 接口行为是否符合需求 MD 中描述的预期

---

## 二、现有能力盘点

| 能力 | 状态 | 关键模块 |
|---|---|---|
| TAPD → MD 抓取 | ✅ 已实现 | `home_task.go` → 异步任务 → `scrape_markdown.go` |
| Playwright 浏览器操作 | ✅ 已实现 | `plw/` 包，支持 click/input/wait/提取/判断 等 15+ 操作 |
| API 执行器 | ✅ 已实现 | `/api/ApiRun`，返回完整响应 |
| API 定义管理 | ✅ 已实现 | `/api/CreateApi`、`/api/Apis`、`/api/ApisDetailByIds` |
| 环境管理 | ✅ 已实现 | `/api/CollectionEnvs`、`/api/CreateCollectionEnv` |
| 浏览器 Session 管理 | ✅ 已实现 | `context_page.go`，BrowserContext 持久化 |
| Agent 远程执行 | ✅ 已实现 | `dtool-agent` 通过 WebSocket 接收并执行任务 |
| 分支变更检测 | ✅ 已实现 | `show-branch-diff` 脚本 |
| 知识片段/MD 存储 | ✅ 已实现 | `memory/service.go`，文件系统 + YAML frontmatter |

---

## 三、整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        AI 编排层（新增）                          │
│  ┌───────────┐  ┌──────────────┐  ┌────────────┐  ┌──────────┐ │
│  │ MD 需求    │  │ 测试计划生成  │  │ 测试执行    │  │ 结果验证  │ │
│  │ 解析器    │→ │ 引擎         │→ │ 引擎        │→ │ & 报告   │ │
│  └───────────┘  └──────────────┘  └────────────┘  └──────────┘ │
└──────────┬──────────────┬──────────────┬───────────────────────┘
           │              │              │
     ┌─────▼─────┐  ┌─────▼─────┐  ┌────▼────┐
     │ 知识片段   │  │ API Runner │  │ Playwright│
     │ /MD 存储   │  │ /api/ApiRun│  │ Smart Link│
     │ (已有)     │  │ (已有)     │  │ + Agent   │
     └───────────┘  └───────────┘  └──────────┘
```

---

## 四、分阶段实施方案

### Phase 1：API 级别自动化测试

**定位：** 纯 API 测试，不涉及浏览器操作，复用现有接口执行能力。

#### 4.1.1 流程

```
1. 获取需求 MD（从知识片段或直接读取）
2. AI 解析 MD，提取：
   - 涉及的接口列表（路径、方法）
   - 每个接口的请求参数（正常值、边界值、异常值）
   - 每个接口的期望返回（状态码、字段结构、业务规则）
3. AI 生成测试用例集（JSON 结构）
4. 调用 /api/ApiRun 逐条执行测试用例
5. AI 对比实际返回 vs 期望返回，判定通过/失败
6. 汇总生成测试报告
```

#### 4.1.2 测试用例结构设计

```json
{
  "test_plan_name": "XX需求接口测试",
  "source_md": "fragments/2026/2026-04/xxx.md",
  "api_base_url": "http://localhost:8080",
  "test_cases": [
    {
      "name": "创建用户-正常流程",
      "api_uri": "/api/user/create",
      "method": "POST",
      "content_type": "application/json",
      "params": {
        "username": "test_user",
        "email": "test@example.com"
      },
      "assertions": [
        {"type": "status_code", "expected": 200},
        {"type": "json_path", "path": "code", "expected": 0},
        {"type": "json_path", "path": "data.id", "expected_type": "number"},
        {"type": "json_path", "path": "data.username", "expected": "test_user"}
      ],
      "category": "positive"
    },
    {
      "name": "创建用户-用户名已存在",
      "api_uri": "/api/user/create",
      "method": "POST",
      "content_type": "application/json",
      "params": {
        "username": "exist_user",
        "email": "test@example.com"
      },
      "assertions": [
        {"type": "json_path", "path": "code", "expected": 10001}
      ],
      "category": "negative"
    },
    {
      "name": "创建用户-缺少必填参数",
      "api_uri": "/api/user/create",
      "method": "POST",
      "content_type": "application/json",
      "params": {
        "email": "test@example.com"
      },
      "assertions": [
        {"type": "json_path", "path": "code", "expected": 400}
      ],
      "category": "boundary"
    }
  ]
}
```

#### 4.1.3 断言类型

| 断言类型 | 说明 | 示例 |
|---|---|---|
| `status_code` | HTTP 状态码 | `{"type": "status_code", "expected": 200}` |
| `json_path` | JSON 字段值匹配 | `{"type": "json_path", "path": "code", "expected": 0}` |
| `json_type` | JSON 字段类型检查 | `{"type": "json_type", "path": "data.id", "expected": "number"}` |
| `json_contains` | JSON 包含指定字段 | `{"type": "json_contains", "path": "data.list"}` |
| `json_not_null` | 字段不为空 | `{"type": "json_not_null", "path": "data.token"}` |
| `response_time` | 响应时间 | `{"type": "response_time", "expected_ms": 3000}` |

#### 4.1.4 测试报告结构

```json
{
  "test_plan_name": "XX需求接口测试",
  "run_time": "2026-04-28T10:30:00+08:00",
  "total": 10,
  "passed": 8,
  "failed": 1,
  "error": 1,
  "results": [
    {
      "name": "创建用户-正常流程",
      "status": "passed",
      "duration_ms": 120,
      "assertions": [
        {"type": "status_code", "expected": 200, "actual": 200, "passed": true},
        {"type": "json_path", "path": "code", "expected": 0, "actual": 0, "passed": true}
      ]
    },
    {
      "name": "创建用户-失败示例",
      "status": "failed",
      "duration_ms": 85,
      "assertions": [
        {"type": "json_path", "path": "code", "expected": 10001, "actual": 0, "passed": false}
      ],
      "actual_response": {"code": 0, "msg": "success", "data": {}}
    }
  ]
}
```

#### 4.1.5 可行性评估

| 维度 | 评分 | 说明 |
|---|---|---|
| 技术可行性 | ★★★★★ | 所有底层能力已就绪 |
| 实现难度 | ★★☆☆☆ | 主要是 AI 编排逻辑 |
| 新增代码量 | 少 | 主要是测试报告结构 + AI 提示词 |
| 覆盖场景 | API 级别 | 无法覆盖 UI 交互 |

---

### Phase 2：Playwright 浏览器端到端测试

**定位：** 补充 UI 层面的端到端测试，覆盖需要浏览器交互的场景。

#### 4.2.1 流程

```
1. 获取需求 MD
2. AI 解析 MD 中涉及 UI 交互的测试场景
3. AI 生成 Smart Link Process 配置（JSON）
4. 加载已有的 Session 登录态（或执行登录流程）
5. 通过 dtool-agent 执行 Playwright 操作
6. 提取页面内容/API 响应，验证是否符合 MD 期望
```

#### 4.2.2 Session/登录态管理

**方案 A：复用 Smart Link 已有 Session**

```
已有的 Smart Link 在执行时会创建 BrowserContext 并保持登录态
→ 新的测试任务直接复用该 Context
→ 需要新增 API：获取可用 Context 列表 / 指定 Context 执行任务
```

**方案 B：基于 storageState 的 Session 导出/加载**

```
1. 用户手动或自动登录目标系统
2. 调用 Playwright API 导出 storageState（cookies + localStorage）
3. 保存为 JSON 文件到指定目录
4. 测试执行时加载 storageState 创建新 Context
5. 测试结束后可选择更新 storageState
```

需要新增的接口：

| 接口 | 说明 |
|---|---|
| `/api/SmartLinkExportSession` | 导出指定 Context 的 storageState |
| `/api/SmartLinkImportSession` | 从 storageState 文件创建新 Context |
| `/api/SmartLinkSessionList` | 列出已保存的 Session 文件 |

#### 4.2.3 测试用例结构

```json
{
  "name": "用户登录后创建订单-E2E测试",
  "session_id": "tapd_logged_in",
  "steps": [
    {"type": "redirect", "url": "/order/create"},
    {"type": "input", "locator": {"role": "textbox", "name": "商品名称"}, "value": "测试商品"},
    {"type": "input", "locator": {"role": "spinbutton", "name": "数量"}, "value": "2"},
    {"type": "click", "locator": {"role": "button", "name": "提交订单"}},
    {"type": "wait_url", "value": "/api/order/create", "wait_seconds": 5},
    {"type": "text_content", "locator": {"css": ".order-result"}, "out_key": "result_text"}
  ],
  "assertions": [
    {"type": "element_exists", "locator": {"css": ".order-success"}, "expected": true},
    {"type": "text_contains", "key": "result_text", "expected": "下单成功"}
  ]
}
```

#### 4.2.4 可行性评估

| 维度 | 评分 | 说明 |
|---|---|---|
| 技术可行性 | ★★★★☆ | Playwright 运行时完整，Session 管理需小幅扩展 |
| 实现难度 | ★★★☆☆ | AI 生成 Process 配置的准确度需要调优 |
| 新增代码量 | 中等 | Session 管理 API + AI 生成配置的提示词 |
| 覆盖场景 | API + UI | 可覆盖完整用户操作链路 |

---

### Phase 3：全自动闭环

**定位：** 从 `tapd_url` 到测试报告的全自动流水线。

#### 4.3.1 完整流程

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ tapd_url  │───→│ TAPD抓取  │───→│ AI解析   │───→│ 分支diff  │───→│ 测试计划  │
│ 任务触发  │    │ → MD     │    │ 需求     │    │ 检测改动  │    │ 生成     │
└──────────┘    └──────────┘    └──────────┘    └──────────┘    └──────────┘
                                                                      │
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐          │
│ 测试报告  │←───│ 结果验证  │←───│ 测试执行  │←───│ 环境准备  │←─────────┘
│ 生成     │    │ & 断言   │    │ (API+UI) │    │ Session  │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
```

#### 4.3.2 任务编排引擎

基于现有的异步任务系统（`async_task`）扩展：

```go
// 测试流水线任务类型
const (
    AsyncTaskTestPlanGenerate  = "test_plan_generate"   // 生成测试计划
    AsyncTaskTestPlanExecute   = "test_plan_execute"    // 执行测试计划
    AsyncTaskTestPlanReport    = "test_plan_report"      // 生成测试报告
)

// 测试流水线配置
type TestPipeline struct {
    SourceMDPath    string   // 需求 MD 路径
    BranchName      string   // 当前分支
    BaseBranch      string   // 基分支（用于 diff）
    TestTypes       []string // 测试类型：api / e2e
    SessionID       string   // E2E 测试用的 Session ID
    CollectionID    int      // 目标 API 集合 ID
    EnvID           int      // 测试环境 ID
}
```

#### 4.3.3 可行性评估

| 维度 | 评分 | 说明 |
|---|---|---|
| 技术可行性 | ★★★☆☆ | 技术上可行，但 AI 判断准确性需持续优化 |
| 实现难度 | ★★★★☆ | 编排引擎 + 错误恢复 + AI 重试 |
| 新增代码量 | 较多 | 编排引擎 + AI 评估逻辑 + 报告持久化 |
| 覆盖场景 | 全链路 | 从需求到测试的完整自动化 |

---

## 五、关键风险与应对策略

| 风险 | 影响 | 概率 | 应对策略 |
|---|---|---|---|
| MD 需求描述不精确 | AI 无法提取明确的测试条件 | 高 | 定义 MD 模板规范，要求必须包含：接口路径、请求参数、期望返回、业务规则 |
| AI 生成的测试用例不完整 | 遗漏边界场景、误判通过 | 中 | 人工审核测试计划 + 持续优化 AI 提示词 + 建立测试用例模板库 |
| 目标系统登录态过期 | Playwright 操作失败 | 中 | Session 自动检测 + 过期自动触发重新登录流程 |
| API 响应不稳定 | 网络抖动导致误报测试失败 | 低 | 内置重试机制（最多 3 次）+ 响应断言容忍度配置 |
| 测试数据污染 | 测试用例产生脏数据 | 中 | 测试用例标记为测试数据 + 提供清理接口 + 使用独立测试环境 |
| AI 对需求理解偏差 | 生成了错误的断言条件 | 中 | 生成测试计划后人工确认，执行结果异常时标注待人工复核 |

---

## 六、推荐实施路径

```
Phase 1 ──────────────────────────────────────────────
  ┌─ 1. 定义测试用例 JSON 结构
  ├─ 2. 定义测试报告 JSON 结构
  ├─ 3. 编写 AI 提示词：MD → 测试计划
  ├─ 4. 实现测试执行引擎（调用 /api/ApiRun）
  ├─ 5. 实现断言引擎（对比实际 vs 期望）
  └─ 6. 实现测试报告生成
     ↓
Phase 2 ──────────────────────────────────────────────
  ┌─ 1. 新增 Session 导出/加载 API
  ├─ 2. 编写 AI 提示词：MD → Smart Link Process 配置
  ├─ 3. 实现 E2E 测试执行引擎
  └─ 4. 实现 UI 断言（元素存在、文本匹配等）
     ↓
Phase 3 ──────────────────────────────────────────────
  ┌─ 1. 扩展异步任务系统，支持测试流水线
  ├─ 2. 实现任务编排引擎
  ├─ 3. 实现 AI 评估与重试机制
  └─ 4. 实现测试报告持久化与前端展示
```

---

## 七、结论

**方案完全可行。** 项目已具备所有底层能力：

1. **需求获取**：TAPD → MD 管线已完整实现
2. **浏览器自动化**：Playwright + Agent 模式已深度集成
3. **接口测试**：API Runner + 环境管理已就绪
4. **登录态管理**：BrowserContext + storageState 原生支持

**建议从 Phase 1 起步**，纯 API 级别测试，复用 `/api/ApiRun`，价值最大、风险最低、实现最快。
