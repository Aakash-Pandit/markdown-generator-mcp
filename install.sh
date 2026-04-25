#!/bin/sh
set -e

echo "Installing markdown-generator-mcp..."

if ! command -v go >/dev/null 2>&1; then
    echo "Error: Go is required (1.21+). Install from https://go.dev/dl/"
    exit 1
fi

go install github.com/Aakash-Pandit/markdown-generator-mcp@latest

echo "Registering with Claude Code..."
"$(go env GOPATH)/bin/mdgen-mcp" --install
