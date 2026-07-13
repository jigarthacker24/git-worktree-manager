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

func TestAddWorktreeArgs(t *testing.T) {
	tests := []struct {
		name         string
		wtPath       string
		branch       string
		newBranch    bool
		sourceBranch string
		want         []string
	}{
		{
			name:      "existing branch",
			wtPath:    "/repo/feature",
			branch:    "feature",
			newBranch: false,
			want:      []string{"worktree", "add", "/repo/feature", "feature"},
		},
		{
			name:         "new branch from source",
			wtPath:       "/repo/new-wt",
			branch:       "feature/new",
			newBranch:    true,
			sourceBranch: "main",
			want:         []string{"worktree", "add", "-b", "feature/new", "/repo/new-wt", "main"},
		},
		{
			name:      "new branch default start",
			wtPath:    "/repo/new-wt",
			branch:    "feature/new",
			newBranch: true,
			want:      []string{"worktree", "add", "-b", "feature/new", "/repo/new-wt"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := addWorktreeArgs(tc.wtPath, tc.branch, tc.newBranch, tc.sourceBranch)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			}
		})
	}
}
