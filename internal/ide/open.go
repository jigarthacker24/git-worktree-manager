package ide

import (
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func Open(path string, kind Kind) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	switch kind {
	case VSCode:
		return openVSCode(abs)
	case Cursor:
		return openCursor(abs)
	case Claude:
		return openClaude(abs)
	default:
		return fmt.Errorf("unsupported IDE")
	}
}

func openVSCode(abs string) error {
	if err := start("code", abs); err == nil {
		return nil
	}
	if runtime.GOOS == "darwin" {
		if err := start("open", "-a", "Visual Studio Code", abs); err == nil {
			return nil
		}
	}
	return fmt.Errorf("could not open VS Code (install the code shell command or ensure VS Code is installed)")
}

func openCursor(abs string) error {
	if err := start("cursor", abs); err == nil {
		return nil
	}
	if runtime.GOOS == "darwin" {
		if err := start("open", "-a", "Cursor", abs); err == nil {
			return nil
		}
	}
	if runtime.GOOS == "windows" {
		if err := start("cmd", "/C", "start", "", "cursor", abs); err == nil {
			return nil
		}
	}
	return fmt.Errorf("could not open Cursor (install the cursor shell command or ensure Cursor is installed)")
}

func openClaude(abs string) error {
	for _, cli := range claudeCLIPaths() {
		cmd := exec.Command(cli)
		cmd.Dir = abs
		if err := cmd.Start(); err == nil {
			return nil
		}
	}

	if hasClaudeCLIHandler() {
		if err := openURL(claudeDeepLink(abs)); err == nil {
			return nil
		}
	}

	if hasClaudeDesktop() {
		if err := openClaudeDesktop(abs); err == nil {
			return nil
		}
	}

	switch runtime.GOOS {
	case "linux":
		if err := start("xdg-open", claudeDeepLink(abs)); err == nil {
			return nil
		}
	case "windows":
		if err := start("cmd", "/C", "start", "", claudeDeepLink(abs)); err == nil {
			return nil
		}
	}

	return fmt.Errorf("could not open Claude Code (install the claude CLI, Claude Desktop, or the claude-cli:// handler)")
}

func openClaudeDesktop(abs string) error {
	link := claudeDesktopDeepLink(abs)
	switch runtime.GOOS {
	case "darwin":
		// Route the URL to Claude.app explicitly; folder must use literal slashes.
		return start("open", "-a", "Claude", "-u", link)
	case "linux", "windows":
		return openURL(link)
	default:
		return fmt.Errorf("unsupported OS")
	}
}

func claudeDeepLink(abs string) string {
	return "claude-cli://open?cwd=" + url.QueryEscape(abs)
}

func claudeDesktopDeepLink(abs string) string {
	return "claude://code/new?folder=" + encodeClaudeFolder(abs)
}

// encodeClaudeFolder URL-encodes a path for Claude Desktop while keeping slashes literal.
// Fully encoding slashes (%2F) prevents Claude from loading the folder on some versions.
func encodeClaudeFolder(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		parts[i] = url.QueryEscape(part)
	}
	return strings.Join(parts, "/")
}

func openURL(link string) error {
	switch runtime.GOOS {
	case "darwin":
		return start("open", link)
	case "linux":
		return start("xdg-open", link)
	case "windows":
		return start("cmd", "/C", "start", "", link)
	default:
		return fmt.Errorf("unsupported OS")
	}
}
