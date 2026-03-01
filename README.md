### 新环境安装

go env -w GOPROXY=https://goproxy.cn,direct

### gs扩展安装

```sh
go env -w GOPRIVATE=gitee.com
# 更新到最新tag
go get -u gitee.com/Sxiaobai/gs/v2@latest

```

### 更新计划

1. 数据库变更：文件支持按年按月目录分类执行 ok
2. 接口开发：执行后保存最后一次的结果 ok
3. 接口开发：执行时需要等待保存完后再执行，且自动切换到结果 ok
4. 接口开发：接口详情页面，右上角环境变量支持查看 ok
5. 接口开发：支持文件夹详情中显示所有接口文档，支持一键复制 ok
6. 接口开发：支持从curl命令生成接口 ok
7. 输出监控：修复某些情况下丢失数据问题（可能是缓冲区不够，还没到换行） ok
8. 输出监控：优化拦截展示页面，把拦截的标题和次数用表格更清晰展示 ok
9. docker：docker服务列表固定按自然顺序排序 ok
10. 自定义网页：执行逻辑支持复制新增 ok
11. nginx：增加nginx配置文件一览 wait
12. redis：增加10个最近搜索的key显示 ok

### 开发约定

1. 开发索引见 `AGENTS.md`；页面与弹窗风格细则见 `docs/frontend-style-guide.md`

### go run 启动

ConfigFile设置为config/dtool下面的某个文件
```shell
export CGO_ENABLED=1 && go run -ldflags " -s -w" cmd/dtool/main.go --ConfigFile=company
```

### Wails 桌面版启动（保留浏览器模式）

先安装 Wails CLI（仅首次）：
```shell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

直接运行桌面版：
```shell
go run -tags dev ./cmd/dtool_wails --ConfigFile=company
```

说明：
1. 浏览器模式入口仍为 `cmd/dtool/main.go`，不受影响。
2. 桌面版会在窗口内启动并自动跳转到本地服务地址。
3. 当配置 `base.webPath` 不存在时，会自动兜底到项目内 `web/dist`。
4. Wails 直接用 `go run/go build` 时必须带 tags，手动构建建议使用 `production`。
5. 若不能传命令行参数，可用环境变量 `DTOOL_CONFIG_FILE` 指定配置文件名。

### bat启动

将下面的内容保存为xxx.bat放到build目录即可双击运行，ctrl + c结束

```bat
@echo off
REM 打开网页（异步）
start "" "http://localhost:17170/"

REM 同步运行 dtool，保持在当前控制台
dtool.exe --ConfigFile=xxx

REM 可选：运行结束后暂停（方便看最后日志）
pause

```
