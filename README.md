### 启动
双击start.bat启动

### 新环境安装
go env -w GOPROXY=https://goproxy.cn,direct

### 更新计划
1. 数据库变更：文件支持按年按月目录分类执行 ok
2. 接口开发：执行后保存最后一次的结果 ok
3. 接口开发：执行时需要等待保存完后再执行，且自动切换到结果 ok
4. 接口开发：接口详情页面，右上角环境变量鼠标指上去需要显示所有的变量，同时支持弹窗修改 wait
5. 接口开发：支持文件夹详情中显示所有接口文档，支持一键复制 ok
6. 接口开发：支持从curl命令生成接口 ok

```shell
#公司编译
export CGO_ENABLED=1  
export GOARCH=amd64   
export GOOS=windows
# DbPath数据库目录
# DbName数据库文件名 为空的话取服务名
go build -tags timetzdata -ldflags " -s -w" -o ./build/zhima.exe ./cmd/zhima/main.go
#git add ./build/zhima.exe
#git update-index --chmod=+x ./build/zhima.exe
git ls-files --stage ./build/zhima.exe
go build -tags timetzdata -ldflags " -s -w" -o ./build/zhimaPub.exe ./cmd/zhima/main.go
git ls-files --stage ./build/zhimaPub.exe
```

```shell
export CGO_ENABLED=1
go run -ldflags " -s -w" cmd/zhima/main.go
```
