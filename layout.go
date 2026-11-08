package main

import (
	"fmt"
	"strings"
)

// ComposeAndRender builds the terminal canvas and flushes it in a single write.
// - If image is present: Dynamically locks max image width so ALL right lines (including FLT)
//   align with 100% vertical symmetry, regardless of image aspect ratio or line count.
// - If no image is present (invalid/empty path): Renders full system info flush at Column 0.
func ComposeAndRender(cfg Config, imgLines []string, items []InfoItem) {
	// Dynamically determine the maximum visible width among all rendered image lines
	maxImgWidth := 0
	for _, l := range imgLines {
		if w := visibleWidth(l); w > maxImgWidth {
			maxImgWidth = w
		}
	}
	hasImage := maxImgWidth > 0

	// Prepare formatted right lines with full values (no artificial truncation)
	var rightLines []string
	for _, item := range items {
		if item.IsTitle {
			// Title line (e.g. user@hostname)
			rightLines = append(rightLines, ColorCyan+item.Val+ColorReset)
		} else {
			// Module line: KEY ❯ VALUE (Full text)
			keyPrefix := fmt.Sprintf("%s%s %s%s ", ColorCyan, item.Key, ColorReset, Chevron)
			rightLines = append(rightLines, keyPrefix+item.Val)
		}
	}

	var sb strings.Builder
	// 1. Clear screen & reset cursor: \033[2J\033[3J\033[H
	// 2. Hide cursor during draw, disable auto-wrap: \033[?25l\033[?7l
	sb.WriteString("\033[2J\033[3J\033[H\033[?25l\033[?7l")

	if !hasImage {
		// When no image: everything starts flush at Column 0 with no left margin
		for _, line := range rightLines {
			sb.WriteString(line)
			sb.WriteString("\r\n")
		}
	} else {
		// When image is present: 2-column side-by-side layout
		maxHeight := len(rightLines)
		if len(imgLines) > maxHeight {
			maxHeight = len(imgLines)
		}

		emptyLeft := strings.Repeat(" ", maxImgWidth)

		for i := 0; i < maxHeight; i++ {
			var left, right string

			if i < len(imgLines) {
				left = imgLines[i]
				// Guarantee every line reaches exact maxImgWidth
				if vw := visibleWidth(left); vw < maxImgWidth {
					left += strings.Repeat(" ", maxImgWidth-vw)
				}
			} else {
				// Pad missing image rows with exact maxImgWidth spaces
				left = emptyLeft
			}

			if i < len(rightLines) {
				right = rightLines[i]
			}

			sb.WriteString(left)
			sb.WriteString(ColorReset)
			if right != "" {
				sb.WriteString("   ") // 3-space margin between image and info
				sb.WriteString(right)
			}
			sb.WriteString("\r\n")
		}
	}

	// 3. Always restore cursor & re-enable auto-wrap: \033[?25h\033[?7h
	sb.WriteString("\033[?25h\033[?7h")

	// Single atomic write to stdout
	fmt.Print(sb.String())
}
