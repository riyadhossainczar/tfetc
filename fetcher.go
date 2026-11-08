package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// InfoItem represents a formatted system information line.
type InfoItem struct {
	IsTitle bool
	RawKey  string
	Key     string
	Val     string
}

// versionRegex extracts semantic version numbers e.g. "5.3.15", "0.90.1", "3.7"
var versionRegex = regexp.MustCompile(`\b(\d+\.\d+(?:\.\d+)?)\b`)

// knownShells recognizes all common and modern Linux/Unix shells
var knownShells = map[string]bool{
	"bash": true, "zsh": true, "fish": true, "nu": true, "nushell": true,
	"elvish": true, "xonsh": true, "ksh": true, "mksh": true, "csh": true,
	"tcsh": true, "dash": true, "ash": true, "sh": true, "oil": true, "yash": true,
}

// getSystemTitle generates a clean user@hostname title when fastfetch omits it (e.g. on Android Termux)
func getSystemTitle() string {
	user := os.Getenv("USER")
	if user == "" {
		out, err := exec.Command("whoami").Output()
		if err == nil {
			user = strings.TrimSpace(string(out))
		}
	}
	if user == "" {
		user = "user"
	}

	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "localhost"
	}
	if os.Getenv("TERMUX_VERSION") != "" && (host == "" || strings.HasPrefix(host, "android-")) {
		host = "localhost"
	}
	return fmt.Sprintf("%s@%s", user, host)
}

// getRealShell dynamically detects the user's active shell (bash, zsh, nushell, fish, etc.)
// and queries its version, preventing fastfetch from reporting this binary ("tfetc") as the shell.
func getRealShell() string {
	shellName := ""

	// 1. Inspect immediate parent process via /proc/<ppid>/comm
	ppid := os.Getppid()
	commBytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", ppid))
	if err == nil {
		comm := strings.TrimSpace(string(commBytes))
		if knownShells[comm] {
			shellName = comm
		} else if comm != "tfetc" && comm != "sudo" && comm != "su" && comm != "systemd" && comm != "init" && comm != "agy" {
			shellName = comm
		}
	}

	// 2. Fallback to SHELL environment variable
	if shellName == "" {
		if envShell := os.Getenv("SHELL"); envShell != "" {
			shellName = filepath.Base(envShell)
		}
	}

	if shellName == "" {
		shellName = "bash"
	}

	// 3. Query shell version via "<shell> --version"
	version := ""
	out, err := exec.Command(shellName, "--version").Output()
	if err == nil {
		m := versionRegex.FindString(string(out))
		if m != "" {
			version = m
		}
	} else {
		outV, errV := exec.Command(shellName, "-v").Output()
		if errV == nil {
			m := versionRegex.FindString(string(outV))
			if m != "" {
				version = m
			}
		}
	}

	if version != "" {
		return fmt.Sprintf("%s %s", shellName, version)
	}
	return shellName
}

// FetchSystemInfo runs fastfetch with --pipe (raw clean text, no ANSI overhead),
// extracts modules, queries FLT binary, and prepares InfoItems.
func FetchSystemInfo(cfg Config) ([]InfoItem, error) {
	// Call fastfetch in raw pipe mode
	args := []string{"--pipe", "--logo", "none"}
	if cfg.FastfetchConfig != "" {
		args = append(args, "--config", cfg.FastfetchConfig)
	}

	cmd := exec.Command("fastfetch", args...)
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	_ = cmd.Run()

	lines := strings.Split(outBuf.String(), "\n")

	type rawEntry struct {
		isTitle bool
		title   string
		key     string
		val     string
	}

	var rawEntries []rawEntry
	km := NewKeyManager()
	firstLine := true
	hasTitle := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "===") {
			continue
		}

		if firstLine {
			firstLine = false
			// Check if line 1 is a title (contains no ": ")
			if !strings.Contains(trimmed, ": ") {
				rawEntries = append(rawEntries, rawEntry{
					isTitle: true,
					title:   trimmed,
				})
				hasTitle = true
				continue
			}
			// If line 1 contains ": ", fastfetch did NOT emit a title (e.g. on Termux).
			// We will prepend a generated title after parsing.
		}

		parts := strings.SplitN(trimmed, ": ", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])

			// Strip phone OEM prefix if present
			v = strings.TrimPrefix(v, "Infinix ")
			v = strings.TrimPrefix(v, "Infinix")

			// Fix fastfetch reporting parent process (tfetc) as the shell
			cleanK := strings.ToUpper(cleanRawKey(k))
			if cleanK == "SHELL" || cleanK == "SHL" {
				if strings.Contains(v, "tfetc") || v == "tfetc" {
					v = getRealShell()
				}
			}

			km.RegisterRawKey(k)
			rawEntries = append(rawEntries, rawEntry{
				isTitle: false,
				key:     k,
				val:     v,
			})
		}
	}

	// If no title was emitted by fastfetch (e.g. on Android Termux), inject generated user@host
	if !hasTitle {
		rawEntries = append([]rawEntry{{
			isTitle: true,
			title:   getSystemTitle(),
		}}, rawEntries...)
	}

	// Resolve FLT custom binary / timestamp
	fltVal := getFltValue(cfg.FltBinary)
	km.RegisterRawKey("FLT")
	rawEntries = append(rawEntries, rawEntry{
		isTitle: false,
		key:     "FLT",
		val:     fltVal,
	})

	// Build resolved InfoItems
	var items []InfoItem
	for _, entry := range rawEntries {
		if entry.isTitle {
			items = append(items, InfoItem{
				IsTitle: true,
				Val:     entry.title,
			})
		} else {
			shortKey := km.ResolveKey(entry.key)
			items = append(items, InfoItem{
				IsTitle: false,
				RawKey:  entry.key,
				Key:     shortKey,
				Val:     entry.val,
			})
		}
	}

	return items, nil
}

// getFltValue executes the custom FLT binary or falls back to standard date format.
func getFltValue(binaryPath string) string {
	if binaryPath != "" {
		resolvedPath := binaryPath
		if !filepath.IsAbs(resolvedPath) && !strings.Contains(resolvedPath, string(filepath.Separator)) {
			if lp, err := exec.LookPath(resolvedPath); err == nil {
				resolvedPath = lp
			}
		}
		if fi, err := os.Stat(resolvedPath); err == nil && (fi.Mode()&0111 != 0) {
			out, err := exec.Command(resolvedPath).Output()
			if err == nil {
				res := strings.TrimSpace(string(out))
				if res != "" {
					return res
				}
			}
		}
	}
	return time.Now().Format("03:04 PM | Mon, 02 Jan 2006")
}
