package ide

import (
	"os"
	"path/filepath"
	"runtime"
)

func Detect() Availability {
	cli := hasClaudeCodeRuntime()
	return Availability{
		VSCode:           detectVSCode(),
		Cursor:           detectCursor(),
		Claude:           cli,
		ClaudeDesktopApp: hasClaudeDesktop(),
	}
}

func detectVSCode() bool {
	if hasCLI("code") {
		return true
	}
	if runtime.GOOS == "darwin" {
		return appExists("/Applications/Visual Studio Code.app")
	}
	return false
}

func detectCursor() bool {
	if hasCLI("cursor") {
		return true
	}
	if runtime.GOOS == "darwin" {
		return appExists("/Applications/Cursor.app")
	}
	return false
}

// hasClaudeCodeRuntime reports whether Claude Code can open a worktree (CLI or URL handler).
// Claude.app alone is the chat client; desktop Code requires Pro/Max and is not counted here.
func hasClaudeCodeRuntime() bool {
	if len(claudeCLIPaths()) > 0 {
		return true
	}
	if hasClaudeCLIHandler() {
		return true
	}
	for _, p := range claudeCodeAppPaths() {
		if appExists(p) {
			return true
		}
	}
	return false
}

// claudeCLIPaths returns likely claude binary locations (PATH and common install dirs).
// GUI apps often launch without ~/.local/bin on PATH.
func claudeCLIPaths() []string {
	seen := make(map[string]struct{})
	var paths []string
	add := func(p string) {
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		if !isExecutable(p) {
			return
		}
		seen[p] = struct{}{}
		paths = append(paths, p)
	}

	if p, err := execLookPath("claude"); err == nil {
		add(p)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return paths
	}
	add(filepath.Join(home, ".local", "bin", "claude"))
	add("/opt/homebrew/bin/claude")
	add("/usr/local/bin/claude")
	return paths
}

func claudeCodeAppPaths() []string {
	var paths []string
	if runtime.GOOS == "darwin" {
		if home, err := os.UserHomeDir(); err == nil {
			paths = append(paths,
				filepath.Join(home, "Applications", "Claude Code URL Handler.app"),
				filepath.Join(home, "Applications", "Claude Code.app"),
				filepath.Join(home, ".local", "share", "claude", "ClaudeCode.app"),
			)
		}
	}
	return paths
}

func hasClaudeCLIHandler() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	return appExists(filepath.Join(home, "Applications", "Claude Code URL Handler.app"))
}

func hasClaudeDesktop() bool {
	switch runtime.GOOS {
	case "darwin":
		return appExists("/Applications/Claude.app")
	default:
		return false
	}
}

func hasCLI(name string) bool {
	_, err := execLookPath(name)
	return err == nil
}

func appExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	return info.Mode()&0111 != 0
}
