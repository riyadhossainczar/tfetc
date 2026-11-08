#!/usr/bin/env bash
# ==============================================================================
# tfetc - Uninstaller Script
# ==============================================================================
set -e

RED="\033[1;31m"
GREEN="\033[1;32m"
CYAN="\033[1;36m"
RESET="\033[0m"

echo -e "${CYAN}==> Uninstalling tfetc...${RESET}"

TARGET_PATHS=(
    "/data/data/com.termux/files/usr/bin/tfetc"
    "${PREFIX}/bin/tfetc"
    "/usr/local/bin/tfetc"
    "/usr/bin/tfetc"
)

REMOVED=0

for path in "${TARGET_PATHS[@]}"; do
    if [ -f "$path" ]; then
        if [ -w "$path" ] || [ "$EUID" -eq 0 ]; then
            rm -f "$path"
        elif command -v sudo >/dev/null 2>&1; then
            sudo rm -f "$path"
        else
            echo -e "${RED}[ERROR] Permission denied removing $path. Please run with sudo.${RESET}"
            continue
        fi
        echo -e "${GREEN}==> Removed: ${path}${RESET}"
        REMOVED=1
    fi
done

# Clean cache directory if exists
CACHE_DIR="${HOME}/.cache/tfetc"
if [ -d "$CACHE_DIR" ]; then
    rm -rf "$CACHE_DIR"
    echo -e "${GREEN}==> Cleaned cache: ${CACHE_DIR}${RESET}"
fi

if [ $REMOVED -eq 1 ]; then
    echo -e "${GREEN}tfetc has been successfully uninstalled.${RESET}"
else
    echo -e "${CYAN}No existing installation of tfetc found.${RESET}"
fi
