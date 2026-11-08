#!/usr/bin/env bash
# ==============================================================================
# tfetc - Universal One-Line Installer for Linux and Android Termux
# ==============================================================================
set -e

REPO_OWNER="${TFETC_REPO_OWNER:-riyadhossainczar}"
REPO_NAME="tfetc"
REPO_BRANCH="${TFETC_REPO_BRANCH:-main}"
RAW_BASE="https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/${REPO_BRANCH}"

# ANSI Colors
GREEN="\033[1;32m"
CYAN="\033[1;36m"
YELLOW="\033[1;33m"
RED="\033[1;31m"
RESET="\033[0m"

echo -e "${CYAN}==> Installing tfetc (Universal Terminal Fetcher)...${RESET}"

# 1. Detect Environment & Target Directory
IS_TERMUX=0
if [ -n "$PREFIX" ] && [ -d "$PREFIX/bin" ]; then
    IS_TERMUX=1
    TARGET_DIR="$PREFIX/bin"
    SUDO=""
elif [ -d "/data/data/com.termux/files/usr/bin" ]; then
    IS_TERMUX=1
    TARGET_DIR="/data/data/com.termux/files/usr/bin"
    SUDO=""
else
    TARGET_DIR="/usr/local/bin"
    if [ "$EUID" -ne 0 ]; then
        if command -v sudo >/dev/null 2>&1; then
            SUDO="sudo"
        else
            echo -e "${RED}[ERROR] Root privileges required to write to ${TARGET_DIR}. Run as root or install sudo.${RESET}"
            exit 1
        fi
    else
        SUDO=""
    fi
fi

# 2. Detect Architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)
        TARGET_ARCH="amd64"
        BINARY_NAME="tfetc"
        ;;
    aarch64|arm64)
        TARGET_ARCH="arm64"
        BINARY_NAME="tfetc-arm64"
        ;;
    *)
        echo -e "${RED}[ERROR] Unsupported CPU architecture: ${ARCH}${RESET}"
        exit 1
        ;;
esac

echo -e "    Environment  : $([ $IS_TERMUX -eq 1 ] && echo 'Android Termux' || echo 'Linux')"
echo -e "    Architecture : ${TARGET_ARCH} (${ARCH})"
echo -e "    Target Path  : ${TARGET_DIR}/tfetc"

# 3. Obtain Binary
TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TEMP_DIR"' EXIT
TARGET_FILE="${TEMP_DIR}/tfetc"

DOWNLOAD_SUCCESS=0

# A. If inside cloned repository with Go source
if [ -f "./main.go" ] && command -v go >/dev/null 2>&1; then
    echo -e "${CYAN}==> Building binary from local Go source...${RESET}"
    if go build -ldflags="-s -w" -o "$TARGET_FILE" .; then
        DOWNLOAD_SUCCESS=1
    fi
# B. If inside cloned repository with prebuilt binary
elif [ -f "./${BINARY_NAME}" ]; then
    echo -e "${CYAN}==> Using prebuilt local binary (${BINARY_NAME})...${RESET}"
    if cp "./${BINARY_NAME}" "$TARGET_FILE"; then
        DOWNLOAD_SUCCESS=1
    fi
fi

# C. Download prebuilt binary from GitHub
if [ $DOWNLOAD_SUCCESS -ne 1 ]; then
    DOWNLOAD_URL="${RAW_BASE}/${BINARY_NAME}"
    echo -e "${CYAN}==> Downloading prebuilt ${TARGET_ARCH} binary from GitHub...${RESET}"
    if command -v curl >/dev/null 2>&1; then
        if curl -sSLf "$DOWNLOAD_URL" -o "$TARGET_FILE" 2>/dev/null; then
            DOWNLOAD_SUCCESS=1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if wget -qO "$TARGET_FILE" "$DOWNLOAD_URL" 2>/dev/null; then
            DOWNLOAD_SUCCESS=1
        fi
    fi
fi

# D. Fallback: If prebuilt binary download failed, compile using Go if Go is available
if [ $DOWNLOAD_SUCCESS -ne 1 ] && command -v go >/dev/null 2>&1; then
    echo -e "${YELLOW}==> Prebuilt binary unavailable. Compiling from source via Go...${RESET}"
    if GOBIN="$TEMP_DIR" go install "github.com/${REPO_OWNER}/${REPO_NAME}@latest" 2>/dev/null; then
        if [ -f "${TEMP_DIR}/tfetc" ]; then
            DOWNLOAD_SUCCESS=1
        fi
    fi
fi

# E. Verify binary integrity
if [ $DOWNLOAD_SUCCESS -ne 1 ] || [ ! -s "$TARGET_FILE" ]; then
    echo -e "${RED}[ERROR] Failed to obtain valid tfetc binary for ${TARGET_ARCH}.${RESET}"
    echo -e "        Please ensure the prebuilt binary is published or install Go to compile:"
    if [ $IS_TERMUX -eq 1 ]; then
        echo -e "        Termux: ${CYAN}pkg install golang git make && git clone https://github.com/${REPO_OWNER}/${REPO_NAME}.git && cd ${REPO_NAME} && make install${RESET}"
    else
        echo -e "        Linux:  ${CYAN}sudo apt install golang git make && git clone https://github.com/${REPO_OWNER}/${REPO_NAME}.git && cd ${REPO_NAME} && sudo make install${RESET}"
    fi
    exit 1
fi

chmod +x "$TARGET_FILE"

# 4. Install Binary to Target Path
$SUDO mkdir -p "$TARGET_DIR"
$SUDO cp "$TARGET_FILE" "${TARGET_DIR}/tfetc"
$SUDO chmod 755 "${TARGET_DIR}/tfetc"

echo -e "${GREEN}==> Successfully installed tfetc to ${TARGET_DIR}/tfetc!${RESET}"

# 5. Check Recommended Dependencies
MISSING_DEPS=""
if ! command -v fastfetch >/dev/null 2>&1; then
    MISSING_DEPS="${MISSING_DEPS} fastfetch"
fi
if ! command -v chafa >/dev/null 2>&1; then
    MISSING_DEPS="${MISSING_DEPS} chafa"
fi

if [ -n "$MISSING_DEPS" ]; then
    echo -e "${YELLOW}==> Note: Recommended tools are missing:${MISSING_DEPS}${RESET}"
    if [ $IS_TERMUX -eq 1 ]; then
        echo -e "    Install them on Termux via: ${CYAN}pkg install${MISSING_DEPS}${RESET}"
    elif command -v apt >/dev/null 2>&1; then
        echo -e "    Install them on Debian/Ubuntu via: ${CYAN}sudo apt install${MISSING_DEPS}${RESET}"
    elif command -v pacman >/dev/null 2>&1; then
        echo -e "    Install them on Arch Linux via: ${CYAN}sudo pacman -S${MISSING_DEPS}${RESET}"
    elif command -v dnf >/dev/null 2>&1; then
        echo -e "    Install them on Fedora via: ${CYAN}sudo dnf install${MISSING_DEPS}${RESET}"
    fi
fi

echo -e ""
echo -e "${GREEN}All done!${RESET} You can now run:"
echo -e "  ${CYAN}tfetc${RESET}          # Standard fetch"
echo -e "  ${CYAN}tfetc -n${RESET}       # Text-only fetch (no image, flush Column 0)"
echo -e "  ${CYAN}tfetc help${RESET}     # View all options and environment variables"
