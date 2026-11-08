package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// safeLookPath searches PATH using os.Stat (fstatat) instead of exec.LookPath.
// This completely bypasses faccessat2 (syscall 439), which Android's seccomp filter
// blocks on Linux kernels < 5.8 (triggering SIGSYS: bad system call).
func safeLookPath(file string) string {
	if strings.Contains(file, "/") {
		return file
	}
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		pathEnv = "/data/data/com.termux/files/usr/bin:/usr/local/bin:/usr/bin:/bin"
	}
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			dir = "."
		}
		candidate := filepath.Join(dir, file)
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() && (fi.Mode()&0111 != 0) {
			return candidate
		}
	}

	// Direct Termux prefix fallback if running on Android Termux
	prefix := os.Getenv("PREFIX")
	if prefix == "" && isDir("/data/data/com.termux/files/usr") {
		prefix = "/data/data/com.termux/files/usr"
	}
	if prefix != "" {
		termuxCandidate := filepath.Join(prefix, "bin", file)
		if fi, err := os.Stat(termuxCandidate); err == nil && !fi.IsDir() && (fi.Mode()&0111 != 0) {
			return termuxCandidate
		}
	}

	return file
}

// isDir checks if path is an existing directory using os.Stat.
func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// safeCommand wraps exec.Command, always passing a resolved path with a slash,
// which guarantees Go's exec.Command bypasses lookPath and never invokes faccessat2.
func safeCommand(name string, arg ...string) *exec.Cmd {
	resolved := safeLookPath(name)
	return exec.Command(resolved, arg...)
}
