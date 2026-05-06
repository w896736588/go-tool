---
name: dtool-playwright
description: Use when operating the dtool 自定义网页 / Playwright 模块 and the task involves running a smart-link login flow, receiving a Chromium userDataDir, and then using native Playwright launchPersistentContext(userDataDir) to take over the logged-in profile locally.
---

# dtool Playwright 技能

- 提供 dtool 自定义网页登录准备接口的使用说明，适用于“先调用 `/api/ai/browser/session/open` 完成登录，再把 `userDataDir` 交给 AI 原生 Playwright 接管”的场景。
- 这个 skill 不再使用 `browser_session_id + action/pages/close` 模型。
- `dtool-playwright` 不在 Skill 列表中，使用时直接内联 Python 调用其 API，Windows 路径用 `r'...'` 原始字符串。

## 强制约束

1. 调用接口前，必须向用户确认以下信息：
   - **请求地址**（`base_url`）：dtool 服务完整地址，例如 `http://127.0.0.1:17170`
   - **Token**：认证令牌，放在请求头 `Token` 中
   - **smart_link_id**：自定义网页配置 ID
   - **label**：自定义网页里目标链接的 label
   - **账号信息**：如果该链接依赖账号，确认使用哪个账号，传 `id` 或 `user_name`
2. 所有请求统一使用 `POST`，`Content-Type: application/json; charset=utf-8`。
3. 统一使用 Python 脚本发送请求，避免 bash/PowerShell 编码问题。
4. 接口执行完成后，服务端会关闭准备阶段浏览器，再返回 `userDataDir`，AI 后续必须自己用原生 Playwright 重新接管。
5. 如果目标 smart-link 配置为“不保存用户数据”，该接口不能返回可复用目录，AI 不能继续原生接管。

## 接口说明

### `POST /api/ai/browser/session/open`

作用：

- 按 `smart_link_id + label + 可选账号` 执行自定义网页流程
- 使用 Chromium 持久化目录完成登录态准备
- 关闭准备阶段浏览器
- 返回 `userDataDir`、账号信息、站点信息、自定义网页标识，供 AI 直接使用原生 Playwright

请求体：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `smart_link_id` | int | 是 | 自定义网页配置 ID |
| `label` | string | 是 | 自定义网页中的链接 label |
| `account` | object | 否 | 账号对象，可传 `{"id": 1}` 或 `{"user_name": "tester"}` |
| `open_type` | int | 否 | 打开方式，`0` 表示沿用配置值 |
| `reuse_if_open` | bool | 否 | 兼容保留字段，默认 `true` |

请求示例：

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

响应重点字段：

| 字段 | 说明 |
|---|---|
| `browser_type` | 固定为 `chromium` |
| `source_browser_closed` | 准备阶段浏览器是否已关闭 |
| `user_data_dir` | AI 后续原生 Playwright 要使用的目录 |
| `user_data_index` | 数据目录索引 |
| `smart_link.id` | 自定义网页配置 ID |
| `smart_link.label` | 链接 label |
| `site.domain` | 站点域名 |
| `site.url` | 打开的站点 URL |
| `account.id` | 账号 ID |
| `account.user_name` | 账号用户名 |
| `native_playwright.mode` | 固定为 `launch_persistent_context` |

响应示例：

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

## “二次附着”说明

二次附着指的是：

- 一个 Chromium 进程已经占用了某个 `userDataDir`
- 另一个 Playwright / Chromium 进程又试图用同一个目录启动

这通常会冲突，所以当前设计里服务端会先关闭准备阶段浏览器，再把目录交给 AI，避免目录仍被占用。

## AI 推荐工作流

### 场景 1：让自定义网页负责登录，AI 原生接管

1. 确认 `base_url`、`Token`、`smart_link_id`、`label`
2. 如果链接依赖账号，确认账号 `id` 或 `user_name`
3. 调用 `/api/ai/browser/session/open`
4. 从返回里读取 `user_data_dir`
5. AI 在本地直接用 Playwright Chromium `launchPersistentContext(userDataDir)` 接管
6. 后续所有点击、输入、截图、断言都走原生 Playwright，不再调用 dtool 的 action 接口

### 场景 2：自动化测试

1. 调 `/api/ai/browser/session/open`
2. 读取 `user_data_dir`
3. 在测试代码中 `launchPersistentContext(userDataDir)`
4. 自己写原生 Playwright 测试步骤和断言

## 原生 Playwright 接管示例

Python：

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

Node.js：

```js
const { chromium } = require("playwright");

async function main() {
  const userDataDir = "C:/path/to/profile/4";
  const context = await chromium.launchPersistentContext(userDataDir, {
    headless: false,
    viewport: null,
  });
  const page = context.pages()[0] || await context.newPage();
  await page.goto("https://example.com/dashboard");
  console.log(await page.title());
}

main();
```

## Python 调用脚本

详细脚本见 `scripts/dtool_playwright_api.py`。
