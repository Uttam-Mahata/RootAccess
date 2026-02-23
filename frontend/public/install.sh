#!/usr/bin/env bash

set -e

BINARY_NAME="rootaccess"
INSTALL_DIR="/usr/local/bin"
BASE_URL="https://ctf.rootaccess.live/bin"

echo "Installing RootAccess CLI..."

OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
    Linux) OS="linux" ;;
    Darwin) OS="darwin" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

DOWNLOAD_URL="$BASE_URL/$BINARY_NAME-$OS-$ARCH"

echo "Downloading $DOWNLOAD_URL..."

# In a real scenario, we would download the binary. 
# For this setup, we assume the user will build it or it will be hosted.
if ! curl -L -f "$DOWNLOAD_URL" -o "$BINARY_NAME"; then
    echo "Error: Could not download binary from $DOWNLOAD_URL"
    echo "Please ensure the binaries are uploaded to the server."
    exit 1
fi

chmod +x "$BINARY_NAME"

if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
else
    echo "Requesting sudo permissions to install to $INSTALL_DIR..."
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
fi

echo "Installation successful!"
echo ""
echo "Run:"
echo "  rootaccess login"
echo "  rootaccess challenges"
