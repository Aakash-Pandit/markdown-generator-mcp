#!/bin/sh
set -e

echo "Installing markdown-generator-mcp..."

if ! command -v go >/dev/null 2>&1; then
    echo "Error: Go is required (1.21+). Install from https://go.dev/dl/"
    exit 1
fi

go install github.com/Aakash-Pandit/markdown-generator-mcp@latest

echo "Registering with Claude Code..."
BINARY="$(go env GOPATH)/bin/markdown-generator-mcp"
if [ ! -f "$BINARY" ]; then
    echo "Error: binary not found at $BINARY"
    exit 1
fi
"$BINARY" --install
