#!/bin/bash
set -e

REPO="lucasmiguelleite/video-downloader-cli"
BINARY="video-downloader"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    darwin) OS="darwin" ;;
    linux)  OS="linux" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
    x86_64)        ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)             echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' | head -1 | sed 's/.*"v\(.*\)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "Error: could not determine latest version"
    exit 1
fi

EXT="tar.gz"
[ "$OS" = "windows" ] && EXT="zip"

URL="https://github.com/$REPO/releases/download/v${VERSION}/${BINARY}_${VERSION}_${OS}_${ARCH}.${EXT}"

echo "Downloading $BINARY v${VERSION} for $OS/$ARCH..."

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$URL" -o "$TMPDIR/$BINARY.${EXT}"

if [ "$EXT" = "zip" ]; then
    unzip -o "$TMPDIR/$BINARY.${EXT}" -d "$TMPDIR" > /dev/null
else
    tar -xzf "$TMPDIR/$BINARY.${EXT}" -C "$TMPDIR"
fi

INSTALL_DIR="/usr/local/bin"

if [ ! -w "$INSTALL_DIR" ]; then
    echo "Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv "$TMPDIR/$BINARY" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$BINARY"
else
    mv "$TMPDIR/$BINARY" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY"
fi

echo "$BINARY v${VERSION} installed to $INSTALL_DIR/$BINARY"
