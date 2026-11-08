package main

import (
	"fmt"
	"os"
	"sync"
)

const Version = "1.0.0"

func printHelp() {
	helpText := fmt.Sprintf(`tfetc v%s - Ultra-fast, Minimalist & Resilient Terminal System Fetcher in Go

USAGE:
  tfetc [COMMAND / FLAGS]

COMMANDS & FLAGS:
  tfetc              Run system fetch with standard dual-column layout
  tfetc -n           Disable image and display system info flush at Column 0
  tfetc --no-image   Disable image (long flag)
  tfetc help         Display this help message (or -h, --help)
  tfetc version      Display version information (or -v, --version)

ENVIRONMENT VARIABLES:
  TFETC_IMAGE        Path to custom image (.jpg, .png, .svg) or "none" to disable
  TFETC_WIDTH        Custom image width in terminal columns (default: 31)
  TFETC_HEIGHT       Custom image height in terminal rows (default: 28)
  TFETC_FLT_BIN      Path to custom executable binary/script for the FLT line
  TFETC_CONFIG       (Optional) Path to custom fastfetch JSONC config file

EXAMPLES:
  tfetc                            # Run with standard dual-column layout
  tfetc -n                         # Run with no image (Column 0 flush)
  tfetc --no-image                 # Same as above
  tfetc help                       # Show help instructions
  TFETC_WIDTH=22 tfetc             # Adjust image width for narrower mobile screens
  TFETC_IMAGE=none tfetc           # Force no-image mode via environment variable
  TFETC_FLT_BIN=~/bin/stamp tfetc  # Use custom timestamp/session binary
`, Version)
	fmt.Print(helpText)
}

func printVersion() {
	fmt.Printf("tfetc v%s\n", Version)
}

func main() {
	noImageFlag := false

	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help", "help":
			printHelp()
			return
		case "-v", "--version", "version":
			printVersion()
			return
		case "-n", "--no-image":
			noImageFlag = true
		}
	}

	cfg := LoadConfig()
	if noImageFlag {
		cfg.ImagePath = ""
	}

	var (
		wg       sync.WaitGroup
		imgLines []string
		items    []InfoItem
	)

	// Fetch Image and System Info concurrently
	wg.Add(2)

	go func() {
		defer wg.Done()
		imgLines = RenderImage(cfg)
	}()

	go func() {
		defer wg.Done()
		items, _ = FetchSystemInfo(cfg)
	}()

	wg.Wait()

	// Composite into memory canvas and flush to terminal
	ComposeAndRender(cfg, imgLines, items)
}
