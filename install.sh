#!/bin/bash
# smaqit-adk installer script
# Usage:
#   gh release download --repo daisy-tribe/smaqit-adk --pattern install.sh -D /tmp/smaqit-adk-install --clobber && bash /tmp/smaqit-adk-install/install.sh
#
# Requires the GitHub CLI (gh), authenticated with access to daisy-tribe/smaqit-adk —
# interactive `gh auth login`, or a GH_TOKEN/GITHUB_TOKEN environment variable for
# headless use. This repo is private, so unauthenticated curl/raw-content access does
# not work; gh is the sole install-time dependency, with no curl-based fallback.
#
# Options (set as environment variables):
#   SMAQIT_ADK_VERSION=latest      Install latest stable ADK release (default)
#   SMAQIT_ADK_VERSION=prerelease  Install latest ADK pre-release (beta, alpha, etc.)
#   SMAQIT_ADK_VERSION=adk-v1.0.0  Install specific ADK version
#
# Examples:
#   gh release download --repo daisy-tribe/smaqit-adk --pattern install.sh -D /tmp/smaqit-adk-install --clobber && bash /tmp/smaqit-adk-install/install.sh
#   gh release download --repo daisy-tribe/smaqit-adk --pattern install.sh -D /tmp/smaqit-adk-install --clobber && SMAQIT_ADK_VERSION=prerelease bash /tmp/smaqit-adk-install/install.sh
#   gh release download --repo daisy-tribe/smaqit-adk --pattern install.sh -D /tmp/smaqit-adk-install --clobber && SMAQIT_ADK_VERSION=adk-v1.0.0 bash /tmp/smaqit-adk-install/install.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="daisy-tribe/smaqit-adk"
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

# Verify gh CLI is installed and authenticated. Required — REPO is private, so
# unauthenticated curl/raw-content access returns 404 for both the release listing
# and the binary download; there is no curl-based fallback.
check_gh_auth() {
    if ! command -v gh &>/dev/null; then
        error "GitHub CLI (gh) is required but not installed. Install it from https://cli.github.com/, run 'gh auth login' (or set GH_TOKEN/GITHUB_TOKEN), then retry."
    fi

    if ! gh auth status &>/dev/null; then
        error "GitHub CLI (gh) is not authenticated. Run 'gh auth login', or set GH_TOKEN/GITHUB_TOKEN to a token with access to ${REPO}, then retry."
    fi

    info "gh CLI authenticated"
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

# Get latest ADK release version via the gh CLI (authenticated — REPO is private)
get_latest_version() {
    info "Fetching ADK release version..."

    case "$SMAQIT_ADK_VERSION" in
        latest)
            # Get latest ADK stable release (tag starts with adk-v, excludes pre-releases)
            VERSION=$(gh release list --repo "$REPO" --exclude-drafts --exclude-pre-releases --json tagName --jq '[.[] | select(.tagName | test("^adk-v[0-9]"))][0].tagName' 2>/dev/null || echo "")

            # Fallback: if no stable ADK release, get most recent ADK pre-release
            if [ -z "$VERSION" ]; then
                warn "No stable ADK release found, using latest ADK pre-release"
                VERSION=$(gh release list --repo "$REPO" --exclude-drafts --json tagName --jq '[.[] | select(.tagName | test("^adk-v"))][0].tagName' 2>/dev/null || echo "")
            fi
            ;;
        prerelease)
            # Get most recent ADK release (including pre-releases)
            VERSION=$(gh release list --repo "$REPO" --exclude-drafts --json tagName --jq '[.[] | select(.tagName | test("^adk-v"))][0].tagName' 2>/dev/null || echo "")
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

# Download binary via gh release download (authenticated — REPO is private)
download_binary() {
    local binary_name="smaqit-adk_${OS}_${ARCH}"

    if [ "$OS" = "windows" ]; then
        binary_name="${binary_name}.exe"
    fi

    local temp_dir
    temp_dir="$(mktemp -d)"
    TEMP_FILE="${temp_dir}/${binary_name}"

    info "Downloading ${binary_name} from release ${VERSION}..."

    if ! gh release download "$VERSION" --repo "$REPO" --pattern "$binary_name" --dir "$temp_dir" --clobber; then
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

# Trigger the global install (agents, skills, templates, framework) via the
# binary's internal --install-global flag. Referenced inline via
# ${INSTALL_DIR}/smaqit-adk — never through an intermediate variable defined
# in another function's scope, since exactly that class of bug (a $target
# variable that fell out of scope) shipped a broken installer once before in
# the sibling smaqit-extensions project, downloading the binary successfully
# but silently never installing anything globally.
install_global() {
    if ! "${INSTALL_DIR}/smaqit-adk" --install-global; then
        error "Global install failed"
    fi
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

    check_gh_auth
    detect_platform
    get_latest_version
    
    download_binary
    install_binary
    verify_installation
    install_global
    check_path

    echo ""
    info "Installation complete!"
    echo ""
    echo "Installed globally:"
    echo "  smaqit.L0/L1/L2  -> ~/.claude/agents/, ~/.codex/agents/"
    echo "  5 ADK skills     -> ~/.agents/skills/, ~/.claude/skills/"
    echo "  templates/framework -> ~/.agents/smaqit-adk/"
    echo ""
    echo "Get started:"
    echo "  In Claude Code or Codex CLI, say 'create a new agent' or 'create a new skill'"
    echo "  smaqit-adk --help     # View available commands"
    echo ""
}

main
