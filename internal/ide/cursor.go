package ide

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

func OpenInCursor(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if err := start("cursor", abs); err == nil {
		return nil
	}

	switch runtime.GOOS {
	case "darwin":
		if err := start("open", "-a", "Cursor", abs); err == nil {
			return nil
		}
	case "windows":
		if err := start("cmd", "/C", "start", "", "cursor", abs); err == nil {
			return nil
		}
	}

	return fmt.Errorf("could not open Cursor (install the cursor shell command or ensure Cursor is installed)")
}

func start(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Start()
}
