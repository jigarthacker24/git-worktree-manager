package gitops

import (
	"testing"
	"time"
)

func TestOrderWorktreesMainAndPins(t *testing.T) {
	wts := []Worktree{
		{Path: "/repo/main", DirName: "main", Main: true},
		{Path: "/repo/z-last", DirName: "z-last"},
		{Path: "/repo/a-first", DirName: "a-first"},
		{Path: "/repo/pinned", DirName: "pinned"},
	}
	norm := func(p string) string { return p }
	ordered := OrderWorktrees(wts, OrderWorktreesOptions{
		Sort:     SortWorktreesByName,
		Pinned:   []string{"/repo/pinned"},
		NormPath: norm,
	})
	if len(ordered) != 4 {
		t.Fatalf("got %d", len(ordered))
	}
	if !ordered[0].Main || ordered[1].Path != "/repo/pinned" {
		t.Fatalf("main/pin order: %+v", ordered[:2])
	}
	if ordered[2].DirName != "a-first" || ordered[3].DirName != "z-last" {
		t.Fatalf("name sort: %+v", ordered[2:])
	}
}

func TestOrderWorktreesByActivity(t *testing.T) {
	now := time.Now()
	wts := []Worktree{
		{Path: "/repo/main", DirName: "main", Main: true},
		{Path: "/repo/old", DirName: "old"},
		{Path: "/repo/new", DirName: "new"},
	}
	activity := map[string]time.Time{
		"/repo/old": now.Add(-2 * time.Hour),
		"/repo/new": now,
	}
	norm := func(p string) string { return p }
	ordered := OrderWorktrees(wts, OrderWorktreesOptions{
		Sort:     SortWorktreesByActivity,
		NormPath: norm,
		ActivityAt: func(p string) time.Time {
			return activity[norm(p)]
		},
	})
	if ordered[1].DirName != "new" || ordered[2].DirName != "old" {
		t.Fatalf("activity sort: %+v", ordered)
	}
}