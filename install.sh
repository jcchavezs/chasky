#!/bin/bash
# Chasky installer script
# Usage: curl -sSL https://raw.githubusercontent.com/jcchavezs/chasky/main/install.sh | bash

set -e

# Determine OS
OS=$(uname -s)
case "${OS}" in
    Linux*)     OS_NAME=Linux;;
    Darwin*)    OS_NAME=Darwin;;
    MINGW*|MSYS*|CYGWIN*) OS_NAME=Windows;;
    *)          echo "Unsupported operating system: ${OS}"; exit 1;;
esac

# Determine architecture
ARCH=$(uname -m)
case "${ARCH}" in
    x86_64)     ARCH_NAME=x86_64;;
    amd64)      ARCH_NAME=x86_64;;
    i386|i686)  ARCH_NAME=i386;;
    arm64|aarch64) ARCH_NAME=arm64;;
    *)          echo "Unsupported architecture: ${ARCH}"; exit 1;;
esac

# Determine file extension
if [ "${OS_NAME}" = "Windows" ]; then
    EXT="zip"
else
    EXT="tar.gz"
fi

REPO="jcchavezs/chasky"
BINARY_NAME="chasky"

echo "Detecting latest version..."
# Try to get version from GitHub API first
LATEST_VERSION=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "")

# If API fails, try scraping from releases page
if [ -z "${LATEST_VERSION}" ]; then
    LATEST_VERSION=$(curl -sSL "https://github.com/${REPO}/releases/latest" 2>/dev/null | grep -o 'tag/v[0-9]*\.[0-9]*\.[0-9]*' | head -1 | cut -d'/' -f2 || echo "")
fi

if [ -z "${LATEST_VERSION}" ]; then
    echo "Failed to detect latest version. Please check your internet connection."
    exit 1
fi

echo "Latest version: ${LATEST_VERSION}"

# Construct download URL
ARCHIVE_NAME="${BINARY_NAME}_${OS_NAME}_${ARCH_NAME}.${EXT}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${ARCHIVE_NAME}"

echo "Downloading ${ARCHIVE_NAME}..."
TEMP_DIR=$(mktemp -d)
cd "${TEMP_DIR}"

if ! curl -sSL -o "${ARCHIVE_NAME}" "${DOWNLOAD_URL}"; then
    echo "Failed to download ${DOWNLOAD_URL}"
    rm -rf "${TEMP_DIR}"
    exit 1
fi

echo "Extracting..."
if [ "${EXT}" = "zip" ]; then
    unzip -q "${ARCHIVE_NAME}"
else
    tar -xzf "${ARCHIVE_NAME}"
fi

# Determine installation directory
if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
elif [ -w "${HOME}/.local/bin" ]; then
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "${INSTALL_DIR}"
else
    INSTALL_DIR="${HOME}/bin"
    mkdir -p "${INSTALL_DIR}"
fi

echo "Installing to ${INSTALL_DIR}..."
if [ "${OS_NAME}" = "Windows" ]; then
    mv "${BINARY_NAME}.exe" "${INSTALL_DIR}/"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}.exe"
else
    mv "${BINARY_NAME}" "${INSTALL_DIR}/"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
fi

# Cleanup
cd - > /dev/null
rm -rf "${TEMP_DIR}"

echo ""
echo "✓ Chasky ${LATEST_VERSION} installed successfully!"
echo ""

# Check if install dir is in PATH
if echo "${PATH}" | grep -q "${INSTALL_DIR}"; then
    echo "You can now use 'chasky' from anywhere."
else
    echo "⚠️  Make sure ${INSTALL_DIR} is in your PATH."
    echo "   Add this to your shell profile (.bashrc, .zshrc, etc.):"
    echo "   export PATH=\"${INSTALL_DIR}:\$PATH\""
fi

echo ""
echo "Get started by running: chasky --help"
