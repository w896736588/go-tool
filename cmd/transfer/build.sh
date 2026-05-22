#!/bin/bash
set -e
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o transfer.exe .
echo "Build done: transfer.exe"
