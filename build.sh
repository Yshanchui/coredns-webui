#!/bin/bash

# 构建脚本

VERSION=${1:-"1.0.0"}
LDFLAGS="-s -w -X main.Version=$VERSION"

echo "开始构建 CoreDNS Web UI v$VERSION"
echo "================================"

# Linux amd64
echo "构建 Linux amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o dist/coredns-webui-linux-amd64 main.go

# Linux arm64
echo "构建 Linux arm64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o dist/coredns-webui-linux-arm64 main.go

# Linux arm
echo "构建 Linux arm..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags="$LDFLAGS" -o dist/coredns-webui-linux-arm main.go

# Windows amd64
echo "构建 Windows amd64..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o dist/coredns-webui-windows-amd64.exe main.go

# macOS amd64
echo "构建 macOS amd64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o dist/coredns-webui-darwin-amd64 main.go

# macOS arm64
echo "构建 macOS arm64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o dist/coredns-webui-darwin-arm64 main.go

echo ""
echo "构建完成！"
echo "================================"
ls -lh dist/
