package gitops

import "testing"

func TestPreferredBranchDisplay(t *testing.T) {
	opts := []BranchOption{
		{Display: "main", Ref: "main", LocalName: "main"},
		{Display: "origin/fix/task-123", Ref: "origin/fix/task-123", RemoteOnly: true, LocalName: "fix/task-123"},
	}
	if PreferredBranchDisplay(opts, "develop") != "main" {
		t.Fatal("expected first option fallback")
	}
	if PreferredBranchDisplay(opts, "fix/task-123") != "origin/fix/task-123" {
		t.Fatal("expected remote by local name match")
	}
	if PreferredBranchDisplay(opts, "main") != "main" {
		t.Fatal("expected local main")
	}
}

func TestBranchOptionWorktreeAddArgs(t *testing.T) {
	local := BranchOption{Display: "feature", Ref: "feature", LocalName: "feature"}
	got := local.WorktreeAddArgs("/wt")
	want := []string{"worktree", "add", "/wt", "feature"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("local add args: got %v want %v", got, want)
		}
	}

	remote := BranchOption{
		Display: "origin/fix/task-123", Ref: "origin/fix/task-123",
		RemoteOnly: true, LocalName: "fix/task-123",
	}
	got = remote.WorktreeAddArgs("/wt")
	want = []string{"worktree", "add", "-b", "fix/task-123", "/wt", "origin/fix/task-123"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("remote add args: got %v want %v", got, want)
		}
	}
}

func TestAddWorktreeArgsFromSpec(t *testing.T) {
	spec := AddWorktreeSpec{
		NewBranch:     true,
		NewBranchName: "feature/x",
		Source:        BranchOption{Ref: "origin/develop", LocalName: "develop"},
	}
	got := addWorktreeArgsFromSpec("/wt", spec)
	want := []string{"worktree", "add", "-b", "feature/x", "/wt", "origin/develop"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("new branch from remote: got %v want %v", got, want)
		}
	}
}
