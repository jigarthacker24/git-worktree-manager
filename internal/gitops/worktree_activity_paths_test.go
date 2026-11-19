package gitops

import "testing"

func TestPathsNeedingActivity(t *testing.T) {
	wts := []Worktree{
		{Path: "/repo/main", Main: true},
		{Path: "/repo/pinned"},
		{Path: "/repo/a"},
		{Path: "/repo/b"},
	}
	norm := func(p string) string { return p }
	got := PathsNeedingActivity(wts, []string{"/repo/pinned"}, norm)
	if len(got) != 2 {
		t.Fatalf("len got %d", len(got))
	}
	seen := map[string]bool{got[0]: true, got[1]: true}
	if !seen["/repo/a"] || !seen["/repo/b"] {
		t.Fatalf("got %v", got)
	}
	if len(PathsNeedingActivity(wts, nil, norm)) != 3 {
		t.Fatal("expected main skipped only")
	}
}