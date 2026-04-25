#!/bin/sh
set -e

BINARY_NAME="markdown-generator-mcp"
INSTALL_DIR="$HOME/.local/bin"
BINARY="$INSTALL_DIR/$BINARY_NAME"

echo "Uninstalling markdown-generator MCP server..."

echo "  Removing MCP registration from Claude Code..."
claude mcp remove --scope user markdown-generator 2>/dev/null || true

echo "  Removing binary..."
if [ -f "$BINARY" ]; then
  rm "$BINARY"
else
  echo "  Binary not found at $BINARY, skipping."
fi

echo ""
echo "Uninstalled successfully. Restart Claude Code."
