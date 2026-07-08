package gitops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "worktree")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := NormalizePath(nested + string(filepath.Separator))
	if err != nil {
		t.Fatal(err)
	}
	want, err := NormalizePath(nested)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
