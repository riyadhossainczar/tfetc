package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

var ansiRegex = regexp.MustCompile(`\x1b(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])`)

// visibleWidth returns the count of visible runes excluding ANSI escape sequences.
func visibleWidth(s string) int {
	clean := ansiRegex.ReplaceAllString(s, "")
	return utf8.RuneCountInString(clean)
}

// RenderImage returns the ANSI lines of the image using chafa with smart disk caching.
func RenderImage(cfg Config) []string {
	if cfg.ImagePath == "" {
		return nil
	}

	imgInfo, err := os.Stat(cfg.ImagePath)
	if err != nil {
		return nil
	}

	// Determine cache directory (~/.cache/tfetc)
	cacheDir := expandPath("~/.cache/tfetc")
	_ = os.MkdirAll(cacheDir, 0755)

	// Create unique hash key for this image path and dimension
	key := fmt.Sprintf("%s_%d_%d", cfg.ImagePath, cfg.ImageWidth, cfg.ImageHeight)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))[:16]
	cacheFile := filepath.Join(cacheDir, hash+".ans")

	// Check if cache file exists and is newer than the source image
	if cacheInfo, err := os.Stat(cacheFile); err == nil {
		if cacheInfo.ModTime().After(imgInfo.ModTime()) {
			cachedBytes, err := os.ReadFile(cacheFile)
			if err == nil && len(cachedBytes) > 0 {
				return splitAnsiLines(string(cachedBytes))
			}
		}
	}

	// Cache miss or outdated: run chafa once
	sizeArg := fmt.Sprintf("%dx%d", cfg.ImageWidth, cfg.ImageHeight)
	cmd := safeCommand("chafa", cfg.ImagePath, "--size", sizeArg)
	outBytes, err := cmd.Output()
	if err != nil || len(outBytes) == 0 {
		return nil
	}

	_ = os.WriteFile(cacheFile, outBytes, 0644)

	return splitAnsiLines(string(outBytes))
}

func splitAnsiLines(raw string) []string {
	raw = strings.TrimRight(raw, "\r\n")
	rawLines := strings.Split(raw, "\n")
	var lines []string

	for _, l := range rawLines {
		cleaned := strings.TrimRight(l, "\r")
		vw := visibleWidth(cleaned)

		// Discard trailing cursor escape sequences (e.g. \033[?25h) that have 0 visible width
		if vw == 0 {
			continue
		}
		lines = append(lines, cleaned)
	}

	return lines
}
