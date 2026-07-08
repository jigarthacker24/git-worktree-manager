package gitops

import "testing"

func TestParseWorktrees(t *testing.T) {
	input := `worktree /repo/main
HEAD abc123
branch refs/heads/main

worktree /repo/feature
HEAD def456
branch refs/heads/feature
`
	wts := parseWorktrees(input)
	if len(wts) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(wts))
	}
	if !wts[0].Main || wts[0].Branch != "main" {
		t.Fatalf("unexpected main worktree: %+v", wts[0])
	}
	if wts[1].Main || wts[1].Branch != "feature" {
		t.Fatalf("unexpected feature worktree: %+v", wts[1])
	}
	if wts[0].DirName != "main" || wts[1].DirName != "feature" {
		t.Fatalf("unexpected dir names: %+v %+v", wts[0], wts[1])
	}
}
