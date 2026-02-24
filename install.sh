#!/bin/bash
set -euo pipefail

# Sloth Incubator installer
# Usage: curl -sSL https://raw.githubusercontent.com/HungSloth/sloth-incubator/main/install.sh | bash

REPO="HungSloth/sloth-incubator"
BINARY_NAME="incubator"
INSTALL_DIR="${HOME}/.local/bin"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    linux)  OS="linux" ;;
    darwin) OS="darwin" ;;
    *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)             echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Detected: ${OS}/${ARCH}"

# Get latest release tag
LATEST=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
    echo "Failed to fetch latest release"
    exit 1
fi
echo "Latest version: ${LATEST}"

# Download
ASSET_NAME="${BINARY_NAME}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET_NAME}"

echo "Downloading ${DOWNLOAD_URL}..."
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

curl -sSL -o "${TMP_DIR}/${ASSET_NAME}" "$DOWNLOAD_URL"

# Extract
tar -xzf "${TMP_DIR}/${ASSET_NAME}" -C "${TMP_DIR}"

# Install
mkdir -p "${INSTALL_DIR}"
mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

echo "Installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"

# Check PATH
if ! echo "$PATH" | grep -q "${INSTALL_DIR}"; then
    echo ""
    echo "Add ${INSTALL_DIR} to your PATH:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
    echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.) to make it permanent."
fi

# Verify
if command -v "${BINARY_NAME}" &> /dev/null; then
    echo ""
    "${BINARY_NAME}" version
    echo "Installation complete!"
else
    echo ""
    echo "Installation complete! Run '${BINARY_NAME} version' to verify."
fi
