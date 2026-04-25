#!/bin/sh
set -e

REPO="Aakash-Pandit/markdown-generator-mcp"
BINARY_NAME="markdown-generator-mcp"
INSTALL_DIR="$HOME/.local/bin"

echo "Installing markdown-generator MCP server..."

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  darwin) OS="darwin" ;;
  linux)  OS="linux" ;;
  *)
    echo "Error: Unsupported OS: $OS"
    echo "Windows users: download the binary manually from https://github.com/$REPO/releases"
    exit 1
    ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Error: Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Get latest release version
echo "  Fetching latest release..."
LATEST=$(curl -sSf "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' | sed 's/.*"tag_name": *"\(.*\)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Error: Could not fetch latest release. Check https://github.com/$REPO/releases"
  exit 1
fi

echo "  Downloading $BINARY_NAME $LATEST ($OS/$ARCH)..."
URL="https://github.com/$REPO/releases/download/$LATEST/${BINARY_NAME}-${OS}-${ARCH}"
curl -sSfL "$URL" -o "/tmp/$BINARY_NAME"
chmod +x "/tmp/$BINARY_NAME"

# Install to ~/.local/bin (no sudo needed)
mkdir -p "$INSTALL_DIR"
mv "/tmp/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"

# Add to PATH if not already there
add_to_path() {
  SHELL_RC=""
  case "$SHELL" in
    */zsh)  SHELL_RC="$HOME/.zshrc" ;;
    */bash) SHELL_RC="$HOME/.bashrc" ;;
  esac
  if [ -n "$SHELL_RC" ] && ! grep -q "$INSTALL_DIR" "$SHELL_RC" 2>/dev/null; then
    echo "" >> "$SHELL_RC"
    echo "export PATH=\"\$HOME/.local/bin:\$PATH\"" >> "$SHELL_RC"
    echo "  Added $INSTALL_DIR to PATH in $SHELL_RC"
    echo "  Run: source $SHELL_RC"
  fi
}

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) add_to_path ;;
esac

echo "  Registering with Claude Code..."
"$INSTALL_DIR/$BINARY_NAME" --install

echo ""
echo "Installed successfully! Restart Claude Code, then say 'make a markdown'."
