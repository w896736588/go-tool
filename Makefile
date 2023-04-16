

.PHONY:xkf_tool_widows
xkf_tool_widows:
	cd cmd/xkf_tool
	set CGO_ENABLED=0;GOOS=windows;GOARCH=amd64 CC=x86_64-w64-mingw32-gcc&&go build -o ./build/xkf_tool.exe ./cmd/xkf_tool/main.go
	cd build&&git add xkf_tool.exe&&git update-index --chmod=+x xkf_tool.exe&&git ls-files --stage xkf_tool.exe

.PHONY:xkf_tool_linux
xkf_tool_linux:
	cd cmd/xkf_tool&&go mod tidy
	set CGO_ENABLED=0&&set GOARCH=amd64&&set GOOS=linux&&go build -o ./build/xkf_tool -ldflags "-s -w" ./cmd/xkf_tool/main.go
	cd build&&git add xkf_tool&&git update-index --chmod=+x xkf_tool&&git ls-files --stage xkf_tool

.PHONY:make_all
make_all:
	make xkf_tool_widows
	make xkf_tool_linux
