#!/bin/sh
set -e

REPO="ljellevo/checkpoint"
BINARY="checkpoint"
INSTALL_DIR="/usr/local/bin"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  darwin|linux) ;;
  *)
    echo "Unsupported OS: $OS"
    echo "For Windows, download the binary manually from https://github.com/$REPO/releases/latest"
    exit 1
    ;;
esac

ASSET="${BINARY}-${OS}-${ARCH}"
URL="https://github.com/$REPO/releases/latest/download/$ASSET"

echo "Downloading $ASSET..."
curl -fsSL "$URL" -o "/tmp/$BINARY"
chmod +x "/tmp/$BINARY"

echo "Installing to $INSTALL_DIR/$BINARY (may require sudo)..."
if [ -w "$INSTALL_DIR" ]; then
  mv "/tmp/$BINARY" "$INSTALL_DIR/$BINARY"
else
  sudo mv "/tmp/$BINARY" "$INSTALL_DIR/$BINARY"
fi

echo "Installed $BINARY at $(which $BINARY)"
echo "$BINARY should now be available in your PATH, and ready to run"