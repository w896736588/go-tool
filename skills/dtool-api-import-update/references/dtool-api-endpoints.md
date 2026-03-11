# dtool 接口开发模块 API 说明

## 基础信息

- 默认服务地址：`http://localhost:17170`
- 请求方式：除 `ApiBatchImport` 外，均为 `POST + application/json`
- 路由前缀：`/api`

## 1. 查询集合与文件夹树：`/api/Collections`

### 请求

```http
POST /api/Collections
Content-Type: application/json

{}
```

### 关键返回字段（data.list）

- `id`: 集合 ID
- `name`: 集合名称
- `type`: `collection`
- `children`: 文件夹数组
  - `children[i].id`: 文件夹 ID
  - `children[i].name`: 文件夹名称
  - `children[i].type`: `folder`

## 2. 新建/更新文件夹：`/api/CreateDir`

### 创建文件夹（不传 id）

```json
{
  "name": "用户中心",
  "collection_id": 1
}
```

### 更新文件夹（传 id）

```json
{
  "id": 12,
  "name": "用户中心-新版",
  "collection_id": 1
}
```

## 3. 查询文件夹详情（含接口列表）：`/api/FolderDetail`

### 请求

```json
{
  "dir_id": 12
}
```

### 关键返回字段（data.dir.children）

- `id`: 接口 ID
- `name`: 接口名称
- `url`: 接口 URI
- `method`: 请求方法

## 4. 新建/更新接口：`/api/CreateApi`

接口逻辑：
- `id` 不存在或为 0 => 新建接口
- `id` 存在 => 更新接口

### 常用参数（JSON）

- `id`：可选，更新时必填
- `folder_id`：必填，文件夹 ID
- `collection_id`：必填，集合 ID
- `name`：接口名称
- `method`：GET/POST/PUT/DELETE...
- `url`：接口 URI
- `protocol`：`http` 或 `https`
- `desc`：描述
- `headers`：对象，例如 `{ "Content-Type": "application/json" }`
- `query_params`：数组，每项至少建议包含 `field`、`type`、`value`、`description`
- `content_type`：如 `application/json`、`multipart/form-data`
- `body_form`：数组，form-data 参数
- `body_json`：JSON 字符串
- `env_id`：可选，环境 ID
- `response_take`、`take_result`、`take_result_desc`：可选

### 创建示例（按 method + content_type）

#### A. GET（常见查询接口）

```json
{
  "folder_id": 12,
  "collection_id": 1,
  "name": "用户列表",
  "method": "GET",
  "url": "$Url$/v1/users",
  "protocol": "https",
  "desc": "分页查询用户",
  "headers": {"Accept": "application/json"},
  "query_params": [
    {"field": "page", "type": "int", "value": "1", "description": "页码"},
    {"field": "size", "type": "int", "value": "20", "description": "每页数量"},
    {"field": "keyword", "type": "string", "value": "", "description": "关键词"}
  ],
  "content_type": "application/json",
  "body_form": [],
  "body_json": ""
}
```

#### B. POST + application/json

```json
{
  "folder_id": 12,
  "collection_id": 1,
  "name": "用户登录",
  "method": "POST",
  "url": "$Url$/v1/login",
  "protocol": "https",
  "desc": "登录接口",
  "headers": {"Content-Type": "application/json"},
  "query_params": [],
  "content_type": "application/json",
  "body_form": [],
  "body_json": "{\"username\":\"demo\",\"password\":\"123456\"}"
}
```

#### C. POST + application/x-www-form-urlencoded

```json
{
  "folder_id": 12,
  "collection_id": 1,
  "name": "短信验证码登录",
  "method": "POST",
  "url": "$Url$/v1/login/sms",
  "protocol": "https",
  "desc": "表单编码提交",
  "headers": {"Content-Type": "application/x-www-form-urlencoded"},
  "query_params": [],
  "content_type": "application/x-www-form-urlencoded",
  "body_form": [
    {"field": "mobile", "type": "string", "value": "13800000000", "description": "手机号"},
    {"field": "code", "type": "string", "value": "123456", "description": "验证码"}
  ],
  "body_json": ""
}
```

#### D. POST + multipart/form-data（含文件）

```json
{
  "folder_id": 12,
  "collection_id": 1,
  "name": "上传头像",
  "method": "POST",
  "url": "$Url$/v1/user/avatar",
  "protocol": "https",
  "desc": "文件上传",
  "headers": {"Content-Type": "multipart/form-data"},
  "query_params": [],
  "content_type": "multipart/form-data",
  "body_form": [
    {"field": "file", "type": "file", "value": "C:/tmp/avatar.png", "description": "头像文件"},
    {"field": "user_id", "type": "int", "value": "1001", "description": "用户ID"}
  ],
  "body_json": ""
}
```

### 更新示例（含 id，按 method + content_type）

#### A. 更新 GET 接口（只要传 id 即为更新）

```json
{
  "id": 301,
  "folder_id": 12,
  "collection_id": 1,
  "name": "用户列表-更新",
  "method": "GET",
  "url": "$Url$/v1/users",
  "protocol": "https",
  "desc": "增加状态筛选",
  "headers": {"Accept": "application/json"},
  "query_params": [
    {"field": "page", "type": "int", "value": "1", "description": "页码"},
    {"field": "size", "type": "int", "value": "20", "description": "每页数量"},
    {"field": "status", "type": "string", "value": "enabled", "description": "状态"}
  ],
  "content_type": "application/json",
  "body_form": [],
  "body_json": ""
}
```

#### B. 更新 POST + application/json

```json
{
  "id": 201,
  "folder_id": 12,
  "collection_id": 1,
  "name": "用户登录",
  "method": "POST",
  "url": "$Url$/v1/login",
  "protocol": "https",
  "desc": "登录接口-更新",
  "headers": {"Content-Type": "application/json"},
  "query_params": [],
  "content_type": "application/json",
  "body_form": [],
  "body_json": "{\"username\":\"demo\",\"password\":\"new-password\"}"
}
```

#### C. 更新 POST + application/x-www-form-urlencoded

```json
{
  "id": 202,
  "folder_id": 12,
  "collection_id": 1,
  "name": "短信验证码登录",
  "method": "POST",
  "url": "$Url$/v1/login/sms",
  "protocol": "https",
  "desc": "增加渠道参数",
  "headers": {"Content-Type": "application/x-www-form-urlencoded"},
  "query_params": [],
  "content_type": "application/x-www-form-urlencoded",
  "body_form": [
    {"field": "mobile", "type": "string", "value": "13800000000", "description": "手机号"},
    {"field": "code", "type": "string", "value": "123456", "description": "验证码"},
    {"field": "channel", "type": "string", "value": "app", "description": "来源渠道"}
  ],
  "body_json": ""
}
```

#### D. 更新 POST + multipart/form-data

```json
{
  "id": 203,
  "folder_id": 12,
  "collection_id": 1,
  "name": "上传头像",
  "method": "POST",
  "url": "$Url$/v1/user/avatar",
  "protocol": "https",
  "desc": "上传头像并附带裁剪参数",
  "headers": {"Content-Type": "multipart/form-data"},
  "query_params": [],
  "content_type": "multipart/form-data",
  "body_form": [
    {"field": "file", "type": "file", "value": "C:/tmp/avatar.png", "description": "头像文件"},
    {"field": "user_id", "type": "int", "value": "1001", "description": "用户ID"},
    {"field": "crop", "type": "string", "value": "{\"x\":0,\"y\":0,\"w\":256,\"h\":256}", "description": "裁剪参数"}
  ],
  "body_json": ""
}
```

## 5. 批量导入接口：`/api/ApiBatchImport`

请求方式：`POST + multipart/form-data`

表单字段：
- `collection_id`: 集合 ID（可与 JSON 内字段二选一，但建议显式传）
- `json`: 导入 JSON 字符串

JSON 结构：

```json
{
  "collection_id": 1,
  "items": [
    {
      "type": "folder",
      "name": "用户中心",
      "desc": "用户接口",
      "children": [
        {
          "type": "api",
          "name": "用户登录",
          "method": "POST",
          "url": "$Url$/v1/login",
          "protocol": "https",
          "desc": "登录",
          "headers": {"Content-Type": "application/json"},
          "body_json": "{\"username\":\"demo\",\"password\":\"123456\"}"
        }
      ]
    }
  ]
}
```

注意：
- 根节点只允许 `folder`。
- `folder.children` 只允许 `api`。
- 同名 folder 导入时会清空该 folder 下旧接口再写入新接口。

## 6. URI 决策流程（导入或更新）

1. 调 `/api/FolderDetail` 获取目标文件夹已有接口。
2. 对待处理接口做 URI 规范化比较。
3. 若命中现有接口：
   - 使用命中接口 `id` 调 `/api/CreateApi` 执行更新。
4. 若未命中：
   - 不传 `id` 调 `/api/CreateApi` 执行导入。

## 7. cURL 调用示例

### 获取集合

```bash
curl -X POST "http://localhost:17170/api/Collections" \
  -H "Content-Type: application/json" \
  -d "{}"
```

### 新建文件夹

```bash
curl -X POST "http://localhost:17170/api/CreateDir" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"用户中心\",\"collection_id\":1}"
```

### 创建接口

```bash
curl -X POST "http://localhost:17170/api/CreateApi" \
  -H "Content-Type: application/json" \
  -d "{\"folder_id\":12,\"collection_id\":1,\"name\":\"用户登录\",\"method\":\"POST\",\"url\":\"$Url$/v1/login\",\"protocol\":\"https\"}"
```

### 批量导入

```bash
curl -X POST "http://localhost:17170/api/ApiBatchImport" \
  -F "collection_id=1" \
  -F "json={\"collection_id\":1,\"items\":[{\"type\":\"folder\",\"name\":\"用户中心\",\"children\":[]}]}"
```
