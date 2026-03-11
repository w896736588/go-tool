---
name: dtool-api-import-update
description: Use when operating the dtool 接口开发模块 and the task requires guiding users to choose collection and folder (including creating a new folder), then deciding import vs update by URI and executing precise API calls.
---

# dtool API 导入/更新执行技能

## 核心目标

让 AI 在“接口开发”场景中稳定执行以下流程：
1. 先让用户选择集合。
2. 再让用户选择文件夹（必须提供“新建文件夹”选项）。
3. 若用户选“新建文件夹”，继续询问新文件夹名称并创建文件夹。
4. 按 URI 判断接口是否已存在：存在则更新，不存在则导入（创建）。

## 执行顺序（必须）

1. 读取 [接口说明](references/dtool-api-endpoints.md)。
2. 调 `/api/Collections` 获取集合与文件夹树。
3. 让用户选择集合（禁止跳过）。
4. 基于所选集合展示文件夹列表，并额外加入“新建文件夹”选项。
5. 若用户选“新建文件夹”，询问新名称并调用 `/api/CreateDir` 创建。
6. 接收用户要导入/更新的接口清单（至少包含 `uri`，建议包含 `name`、`method`、`desc`、请求参数等）。
7. 调 `/api/FolderDetail` 获取目标文件夹下已有接口列表。
8. 对每个待处理接口按 URI 匹配：
   - 匹配到：调用 `/api/CreateApi` 且携带 `id`，执行更新。
   - 未匹配：调用 `/api/CreateApi` 且不带 `id`，执行创建。
9. 汇总结果：创建数量、更新数量、失败明细。

## URI 匹配规则

1. 默认使用“规范化 URI + method”匹配：
   - 去除前后空格。
   - 协议与域名大小写忽略。
   - 去掉末尾 `/` 后比较。
2. 如果用户明确要求“仅按 URI 比较，不区分 method”，按用户要求执行。
3. 如果同一文件夹内命中多个候选，先暂停并让用户确认目标接口 ID。

## 何时用批量导入

当用户提供的是“完整文件夹 + 子接口”结构，且接受“同名文件夹覆盖式更新（会先清空该文件夹下旧接口）”时，优先使用 `/api/ApiBatchImport`。

调用格式：`multipart/form-data`
- `collection_id`: 集合 ID
- `json`: JSON 字符串（结构见 [批量导入格式](references/dtool-api-endpoints.md#批量导入接口apibatchimport)）

## 交互约束

1. 交互必须使用简体中文。
2. 不得在未确认集合/文件夹前直接写入接口。
3. 涉及覆盖风险（批量导入同名文件夹）必须先提示风险再执行。
4. 接口调用失败时，必须返回失败接口名称、URI、错误信息。

## 可选脚本

需要自动执行时，优先使用 `scripts/sync_api_by_uri.py`：
- 读取 JSON 文件。
- 自动完成“查找集合/文件夹/按 URI 创建或更新”。
- 支持 `--create-folder` 在找不到文件夹时自动新建。
