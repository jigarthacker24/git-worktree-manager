package ide

import "testing"

func TestClaudeDeepLink(t *testing.T) {
	got := claudeDeepLink("/repo/my worktree")
	want := "claude-cli://open?cwd=%2Frepo%2Fmy+worktree"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestClaudeDesktopDeepLink(t *testing.T) {
	got := claudeDesktopDeepLink("/repo/my worktree")
	want := "claude://code/new?folder=/repo/my+worktree"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestEncodeClaudeFolder(t *testing.T) {
	got := encodeClaudeFolder("/Users/me/project")
	want := "/Users/me/project"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
