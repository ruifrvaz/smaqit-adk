#!/bin/bash
# smaqit-adk installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit-adk/main/install.sh | bash
# 
# Options (set as environment variables):
#   SMAQIT_ADK_VERSION=latest      Install latest stable ADK release (default)
#   SMAQIT_ADK_VERSION=prerelease  Install latest ADK pre-release (beta, alpha, etc.)
#   SMAQIT_ADK_VERSION=adk-v1.0.0  Install specific ADK version
#
# Examples:
#   curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit-adk/main/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit-adk/main/install.sh | SMAQIT_ADK_VERSION=prerelease bash
#   curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit-adk/main/install.sh | SMAQIT_ADK_VERSION=adk-v1.0.0 bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="ruifrvaz/smaqit-adk"
INSTALL_DIR="${HOME}/.local/bin"
SMAQIT_ADK_VERSION="${SMAQIT_ADK_VERSION:-latest}"  # Default to latest stable ADK

# Helper functions
info() {
    echo -e "${GREEN}✓${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case "$os" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ;;
        *)
            error "Unsupported operating system: $os"
            ;;
    esac
    
    case "$arch" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $arch"
            ;;
    esac
    
    info "Detected platform: ${OS}/${ARCH}"
}

# Get latest ADK release version from GitHub API
get_latest_version() {
    info "Fetching ADK release version..."
    
    local api_url="https://api.github.com/repos/${REPO}/releases"
    
    case "$SMAQIT_ADK_VERSION" in
        latest)
            # Get latest ADK stable release (tag starts with adk-v, excludes pre-releases)
            VERSION=$(curl -fsSL "$api_url" | grep '"tag_name"' | grep 'adk-v' | grep -v '-' | head -1 | sed -E 's/.*"([^"]+)".*/\1/' || echo "")
            
            # Fallback: if no stable ADK release, get most recent ADK pre-release
            if [ -z "$VERSION" ]; then
                warn "No stable ADK release found, using latest ADK pre-release"
                VERSION=$(curl -fsSL "$api_url" | grep '"tag_name"' | grep 'adk-v' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')
            fi
            ;;
        prerelease)
            # Get most recent ADK release (including pre-releases)
            VERSION=$(curl -fsSL "$api_url" | grep '"tag_name"' | grep 'adk-v' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')
            ;;
        adk-v*.*.*)
            # Use specific ADK version provided
            VERSION="$SMAQIT_ADK_VERSION"
            ;;
        *)
            error "Invalid SMAQIT_ADK_VERSION: $SMAQIT_ADK_VERSION (use 'latest', 'prerelease', or 'adk-vX.Y.Z')"
            ;;
    esac
    
    if [ -z "$VERSION" ]; then
        error "Failed to fetch ADK release version"
    fi
    
    info "Installing ADK version: ${VERSION}"
}

# Download binary
download_binary() {
    local binary_name="smaqit-adk_${OS}_${ARCH}"
    
    if [ "$OS" = "windows" ]; then
        binary_name="${binary_name}.exe"
    fi
    
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/${binary_name}"
    TEMP_FILE="/tmp/smaqit-adk_${VERSION}"
    
    info "Downloading from ${download_url}..."
    
    if ! curl -fsSL -o "$TEMP_FILE" "$download_url"; then
        error "Failed to download ADK binary"
    fi
    
    info "Download complete"
}

# Install binary
install_binary() {
    local target="${INSTALL_DIR}/smaqit-adk"
    
    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"
    
    # Make executable
    chmod +x "$TEMP_FILE"
    
    # Move to install directory
    mv "$TEMP_FILE" "$target"
    
    info "Installed to ${target}"
}

# Verify installation
verify_installation() {
    local target="${INSTALL_DIR}/smaqit-adk"
    
    if ! "$target" --version &>/dev/null; then
        error "Installation verification failed"
    fi
    
    local installed_version=$("$target" --version 2>&1 || echo "unknown")
    info "Verified installation: ${installed_version}"
}

# Check if install directory is in PATH
check_path() {
    if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
        warn "${INSTALL_DIR} is not in your PATH"
        echo ""
        echo "Add to your shell config (~/.bashrc, ~/.zshrc, etc.):"
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        echo ""
        echo "Then reload your shell:"
        echo "  source ~/.bashrc  # or ~/.zshrc"
        echo ""
    fi
}

# Main installation flow
main() {
    echo "smaqit-adk installer"
    echo "===================="
    echo ""
    
    detect_platform
    get_latest_version
    
    download_binary
    install_binary
    verify_installation
    check_path
    
    echo ""
    info "Installation complete!"
    echo ""
    echo "Get started:"
    echo "  smaqit-adk init       # Initialize ADK in your project"
    echo "  smaqit-adk --help     # View available commands"
    echo ""
    echo "Next steps:"
    echo "  1. Fill .github/prompts/smaqit.new-agent.prompt.md with agent requirements"
    echo "  2. Open GitHub Copilot chat and type '/smaqit.L2' to compile your agent"
    echo ""
}

main
