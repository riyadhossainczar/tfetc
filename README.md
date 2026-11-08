# tfetc

<div align="center">

**Ultra-fast, Minimalist & Resilient Terminal System Fetcher in Go**  
*Tailored for Android Termux, Linux Codespaces, and Modern Shells.*

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-blue.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20Android%20Termux-brightgreen.svg)](#prerequisites)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Speed](https://img.shields.io/badge/startup-%7E12ms-orange.svg)](#performance)

</div>

---

## ⚡ Overview

`tfetc` combines the rich system detection of **`fastfetch`** with the graphical capabilities of **`chafa`** into a single, high-performance Go binary. It achieves cold and warm startup speeds of **10–15 milliseconds** using concurrent data fetching, in-memory terminal canvas compositing, and smart ANSI disk caching.

Unlike legacy bash fetch scripts that spawn dozens of subshells and suffer from terminal line-wrapping glitches, `tfetc` runs as a compiled binary with zero external configuration file dependencies.

---

## ✨ Key Features

- **⚡ Blazing Fast (~12ms)**: Concurrently fetches image data and hardware telemetry in separate goroutines. Over 100× faster than shell-based fetch scripts.
- **🖼️ Smart ANSI Caching**: Images are rendered via `chafa` once and cached as raw ANSI text (`~/.cache/tfetc/`). Subsequent launches load the image in `< 0.1ms`.
- **🏷️ Strict 3-Letter Key Engine**: Universal 3-character keys (`SYS`, `HST`, `KRN`, `RAM`, `PKG`, `ENV`, `WND`). Multi-instance modules (like multiple Disks) use Base-36 indexing (`DK1`..`DK9`, `DKA`..`DKZ`), guaranteeing strict 3-character column consistency up to 36 disks.
- **📏 Strict Single-Line Canvas (No Wrap Glitches)**: Disables terminal auto-wrap (`\033[?7l`) during rendering so long strings never wrap around to corrupt the image layout. Extra text clips cleanly off-screen on narrow mobile displays.
- **🚫 Zero-Image Mode (Column 0 Flush)**: If no image is provided, if the image path is invalid, or if invoked with `--no-image` / `TFETC_IMAGE=none`, all system info renders flush at Column 0 with no wasted space or blank indentation.
- **📱 Termux & Cross-Platform Resilience**:
  - Automatically generates `$USER@localhost` (e.g. `u0_a2042@localhost`) on Android Termux where fastfetch omits title lines.
  - Correctly detects active shells (`bash`, `zsh`, `fish`, `nushell`, etc.) and their semantic versions.
  - Full ARM64 build (`tfetc-arm64`) included for Android.
- **🛡️ Zero-Config & Graceful Degradation**: Zero hardcoded directory hunting. If a custom binary or image is missing, `tfetc` falls back gracefully without crashing or throwing terminal errors.

---

## 📦 Prerequisites

To run `tfetc`, install `fastfetch` and `chafa`:

### 1. Android Termux
```bash
pkg update && pkg install fastfetch chafa golang git make -y
```

### 2. Debian / Ubuntu / Linux Mint
```bash
sudo apt update && sudo apt install fastfetch chafa golang git make -y
```

### 3. Arch Linux / Manjaro
```bash
sudo pacman -S fastfetch chafa go git make
```

### 4. Fedora / RHEL
```bash
sudo dnf install fastfetch chafa golang git make -y
```

---

## 🚀 Installation & Building

### 1. Clone Repository
```bash
git clone https://github.com/your-username/tfetc.git
cd tfetc
```

### 2. Build Binaries
```bash
# Build for your current machine (Linux x86_64)
make build

# Cross-compile for Android Termux (ARM64)
make build-arm64

# Build both architectures at once
make all
```

### 3. Install System-Wide (Optional)
```bash
# On Linux:
sudo make install

# On Termux:
make install
```
This installs `tfetc` directly into your `$PATH` (`/usr/local/bin/tfetc` or `$PREFIX/bin/tfetc`).

---

## 💻 Usage & CLI Flags

Run `tfetc` directly from your terminal:

```bash
# Run with default dual-column layout
tfetc

# Run in text-only mode (Column 0 flush, no image)
tfetc --no-image
tfetc -n

# Show help information
tfetc --help
tfetc -h

# Show version
tfetc --version
tfetc -v
```

---

## ⚙️ Configuration & Environment Variables

You do not need to recompile `tfetc` to customize its appearance. Configure it on the fly using standard environment variables:

| Variable | Description | Default | Example |
| :--- | :--- | :--- | :--- |
| `TFETC_IMAGE` | Path to custom image (`.jpg`, `.png`, `.svg`) or `none` to disable | Column 0 (text-only) if missing/unset | `export TFETC_IMAGE=~/pictures/avatar.png`<br>`export TFETC_IMAGE=none` |
| `TFETC_WIDTH` | Custom image column width in characters | `31` | `export TFETC_WIDTH=22` |
| `TFETC_HEIGHT` | Custom image row height in lines | `28` | `export TFETC_HEIGHT=20` |
| `TFETC_FLT_BIN` | Path to custom binary or script for session stamping | Custom binary / System Timestamp | `export TFETC_FLT_BIN=~/bin/custom-stamp` |
| `TFETC_CONFIG` | *(Optional)* Custom `fastfetch` JSONC config path | None (raw pipe mode) | `export TFETC_CONFIG=~/.config/fastfetch/custom.jsonc` |

### One-Liner Inline Examples:
```bash
# Run with narrower image for small mobile screens:
TFETC_WIDTH=22 TFETC_HEIGHT=20 tfetc

# Run with a specific wallpaper:
TFETC_IMAGE=~/wallpapers/cyberpunk.png tfetc

# Run completely without image:
TFETC_IMAGE=none tfetc
# Or simply:
tfetc -n
```

---

## 🐚 Shell Startup Integration

To display `tfetc` automatically whenever you open a new terminal session, add it to your shell configuration:

### In Bash (`~/.bashrc`):
```bash
# Add to the bottom of ~/.bashrc
# Optional customization:
export TFETC_IMAGE=~/my_avatar.png
export TFETC_WIDTH=24

# Launch tfetc
tfetc
```

### In Zsh (`~/.zshrc`):
```bash
# Add to the bottom of ~/.zshrc
tfetc
```

### In Fish (`~/.config/fish/config.fish`):
```fish
if status is-interactive
    tfetc
end
```

---

## 📐 Architecture & Workflow

```text
               ┌────────────────────────────────────────────────────────┐
               │                        tfetc                           │
               └────────────────────────────────────────────────────────┘
                                   │
              ┌────────────────────┴────────────────────┐
              ▼                                         ▼
     [ Goroutine 1: Image ]                  [ Goroutine 2: Telemetry ]
   Checks ~/.cache/tfetc/                     fastfetch --pipe --logo none
   Miss? Run chafa once & cache               Raw pipe parse (0 ANSI overhead)
              │                                         │
              │                               Detect Host & Shell version
              │                               3-Letter Key Shortener (Base-36)
              │                                         │
              └────────────────────┬────────────────────┘
                                   ▼
                   [ In-Memory Canvas Compositor ]
                   Dynamic maxImgWidth symmetry padding
                   Disable auto-wrap (\033[?7l)
                                   ▼
                   [ Single Atomic Stdout Write ]
                   Instantaneous ~12ms terminal render
```

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
