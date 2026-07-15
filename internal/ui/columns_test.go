package ui

import (
	"testing"

	"github.com/jigarthacker24/git-worktree-manager/internal/gitops"
)

func TestWorktreeBranchLabel(t *testing.T) {
	if got := WorktreeBranchLabel(gitops.Worktree{Branch: "dev", Main: true}); got != "dev · main" {
		t.Fatalf("got %q", got)
	}
	if got := WorktreeBranchLabel(gitops.Worktree{}); got != "(detached)" {
		t.Fatalf("got %q", got)
	}
}
