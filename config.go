package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Default paths and parameters. Edit here or override via environment variables.
const (
	// Default image locations
	DefaultImagePath = "/workspaces/Serein-SE/store-global.location/specifier.local-termux.pev/mod_pic/t6.jpg"

	// Chafa image rendering dimensions
	DefaultImageWidth  = 31
	DefaultImageHeight = 26

	// External FLT binary location (outputs custom session/device stamp)
	DefaultFltBinary = "/workspaces/Serein-SE/store-global.location/specifier.local-termux.pev/mod_bin/main-amd64"

	// ANSI Color Palette
	ColorCyan  = "\033[1;36m"
	ColorReset = "\033[0m"
	ColorDim   = "\033[90m"
	Chevron    = "❯"
)

// Config holds runtime configuration resolved from defaults and environment variables.
type Config struct {
	ImagePath       string
	ImageWidth      int
	ImageHeight     int
	FltBinary       string
	FastfetchConfig string // Optional: only if user explicitly sets TFETC_CONFIG
}

// expandPath expands ~ and resolves absolute path.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	} else if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	}
	return path
}

// LoadConfig resolves paths with environment variable overrides.
func LoadConfig() Config {
	cfg := Config{
		ImagePath:   DefaultImagePath,
		ImageWidth:  DefaultImageWidth,
		ImageHeight: DefaultImageHeight,
		FltBinary:   DefaultFltBinary,
	}

	// Image path overrides
	if envImg, set := os.LookupEnv("TFETC_IMAGE"); set {
		if envImg == "none" || envImg == "0" || envImg == "off" || envImg == "" {
			cfg.ImagePath = ""
		} else {
			cfg.ImagePath = expandPath(envImg)
		}
	}

	// Image dimensions overrides
	if envW := os.Getenv("TFETC_WIDTH"); envW != "" {
		if w, err := strconv.Atoi(envW); err == nil && w > 0 {
			cfg.ImageWidth = w
		}
	}
	if envH := os.Getenv("TFETC_HEIGHT"); envH != "" {
		if h, err := strconv.Atoi(envH); err == nil && h > 0 {
			cfg.ImageHeight = h
		}
	}

	// Custom FLT binary override
	if envFlt := os.Getenv("TFETC_FLT_BIN"); envFlt != "" {
		cfg.FltBinary = expandPath(envFlt)
	}

	// Optional: Custom fastfetch config file ONLY if explicitly provided by user
	if envCfg := os.Getenv("TFETC_CONFIG"); envCfg != "" {
		cfg.FastfetchConfig = expandPath(envCfg)
	}

	// If the image path does not exist on disk, clear it so everything renders flush at Column 0
	if cfg.ImagePath != "" {
		if _, err := os.Stat(cfg.ImagePath); os.IsNotExist(err) {
			cfg.ImagePath = ""
		}
	}

	return cfg
}
