@echo off
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o transfer.exe .
echo Build done: transfer.exe
