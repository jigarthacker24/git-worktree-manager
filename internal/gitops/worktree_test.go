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

func TestPlanWorktreeMoves(t *testing.T) {
	base := "/projects/worktrees"
	worktrees := []Worktree{
		{Path: "/repo/main", DirName: "main", Main: true},
		{Path: "/projects/worktrees/feature-a", DirName: "feature-a"},
		{Path: "/elsewhere/bugfix", DirName: "bugfix"},
		{Path: "/tmp/hotfix", DirName: "hotfix"},
	}

	moves, err := PlanWorktreeMoves(worktrees, base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(moves) != 2 {
		t.Fatalf("expected 2 moves, got %d: %+v", len(moves), moves)
	}
	if moves[0].From != "/elsewhere/bugfix" || moves[0].To != "/projects/worktrees/bugfix" {
		t.Fatalf("unexpected first move: %+v", moves[0])
	}
	if moves[1].From != "/tmp/hotfix" || moves[1].To != "/projects/worktrees/hotfix" {
		t.Fatalf("unexpected second move: %+v", moves[1])
	}
}

func TestPlanWorktreeMovesCollision(t *testing.T) {
	worktrees := []Worktree{
		{Path: "/a/one", DirName: "wt"},
		{Path: "/b/two", DirName: "wt"},
	}
	_, err := PlanWorktreeMoves(worktrees, "/dest")
	if err == nil {
		t.Fatal("expected collision error")
	}
}

func TestMoveWorktreeArgs(t *testing.T) {
	got := moveWorktreeArgs("/old/wt", "/new/wt", false)
	want := []string{"worktree", "move", "/old/wt", "/new/wt"}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	got = moveWorktreeArgs("/old/wt", "/new/wt", true)
	want = []string{"worktree", "move", "--force", "/old/wt", "/new/wt"}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestAddWorktreeArgsFromSpecExisting(t *testing.T) {
	spec := AddWorktreeSpec{
		Existing: BranchOption{Ref: "feature", LocalName: "feature"},
	}
	got := addWorktreeArgsFromSpec("/repo/feature", spec)
	want := []string{"worktree", "add", "/repo/feature", "feature"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
