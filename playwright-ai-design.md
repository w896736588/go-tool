# Playwright AI 开放能力设计

## 1. 设计目标

新的目标不再是给 AI 提供一组 `/action/pages/close` 会话接口，而是把职责拆开：

1. dtool 自定义网页负责登录和登录态准备
2. 接口返回 Chromium 的 `userDataDir`
3. 准备阶段浏览器关闭
4. AI 自己用原生 Playwright `launchPersistentContext(userDataDir)` 接管

这样做的核心价值：

- 保留现有 smart-link、账号体系、自定义网页登录流程
- 不再维护大量浏览器 action 接口
- AI 可以直接使用原生 Playwright 全能力
- 自动化测试、复杂调试、多步交互都更自然

## 2. 为什么改成这种模型

之前的 `browser_session_id + action` 模型虽然可控，但长期成本高：

- 后端需要持续维护 click、fill、wait、assert、download 等大量动作
- AI 仍然受限于接口设计边界
- 一旦场景复杂，接口参数和返回结构会快速膨胀

而“返回 `userDataDir` 给 AI 原生 Playwright”更符合职责分层：

- 服务端只做自己最擅长的登录准备
- AI 直接用 Playwright 原生 API 做后续操作

## 3. 关键约束

### 3.1 浏览器类型

- 固定使用 `chromium`

### 3.2 准备阶段必须关闭浏览器

接口完成后必须关闭准备阶段浏览器，再返回 `userDataDir`。

原因：

- 如果原浏览器还占着同一个 profile 目录，AI 再用原生 Playwright 启动同目录时容易冲突
- 这类冲突就是“二次附着”

### 3.3 必须返回上下文信息

接口不应只返回一个裸目录路径，还要同时返回：

- 该目录对应哪个账号
- 对应哪个站点
- 对应哪个 `smart_link`

这样 AI 才知道自己接管的到底是什么登录态。

### 3.4 不支持无痕目录

如果某个 smart-link 配置为“不保存用户数据”，那就没有稳定可复用的 profile 目录，不能走这套模式。

## 4. 对外接口

保留：

- `POST /api/ai/browser/session/open`

移除：

- `POST /api/ai/browser/session/action`
- `POST /api/ai/browser/session/pages`
- `POST /api/ai/browser/session/close`

## 5. open 接口语义

### 请求

```json
{
  "smart_link_id": 12,
  "label": "登录后首页",
  "account": {
    "user_name": "tester"
  },
  "open_type": 0,
  "reuse_if_open": true
}
```

### 处理流程

1. 根据 `smart_link_id + label + account` 解析运行参数
2. 要求使用保存用户数据的持久化 Chromium context
3. 打开或复用该目录对应的 context
4. 执行自定义网页登录流程
5. 获取实际使用的 `userDataDir`
6. 收集账号、站点、自定义网页信息
7. 关闭准备阶段浏览器
8. 返回给 AI

## 6. 响应结构

```json
{
  "browser_type": "chromium",
  "source_browser_closed": true,
  "native_playwright": {
    "mode": "launch_persistent_context",
    "user_data_dir": "C:/path/to/profile/4"
  },
  "user_data_dir": "C:/path/to/profile/4",
  "user_data_index": 4,
  "smart_link": {
    "id": 12,
    "label": "登录后首页"
  },
  "site": {
    "domain": "example.com",
    "url": "https://example.com/dashboard"
  },
  "account": {
    "id": 3,
    "user_name": "tester",
    "account_key": "account_id_3"
  },
  "current_page": {
    "url": "https://example.com/dashboard",
    "title": "控制台"
  }
}
```

## 7. AI 使用方式

AI 侧直接使用原生 Playwright。

Python 示例：

```python
from playwright.sync_api import sync_playwright

user_data_dir = r"C:\path\to\profile\4"

with sync_playwright() as p:
    context = p.chromium.launch_persistent_context(
        user_data_dir=user_data_dir,
        headless=False,
        no_viewport=True,
    )
    page = context.pages[0] if context.pages else context.new_page()
    page.goto("https://example.com/dashboard")
    print(page.title())
```

## 8. 当前实现约定

- 只保留 `/api/ai/browser/session/open`
- 使用 Chromium 持久化目录
- 接口完成后关闭准备阶段浏览器
- 返回 `userDataDir + account + site + smart_link`
- 返回结果里附带使用提示，明确告诉 AI 该如何接管目录

## 9. 结论

这套新设计比之前的 session/action 模型更轻：

- dtool 负责登录
- AI 负责原生 Playwright 操作

它更适合你现在的目标，也更适合后续自动化测试和复杂网页操作扩展。
