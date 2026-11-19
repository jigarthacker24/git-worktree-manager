package gitops

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkOrderWorktrees_ByName(b *testing.B) {
	wts := benchWorktrees(100)
	opt := OrderWorktreesOptions{
		Sort:     SortWorktreesByName,
		NormPath: func(p string) string { return p },
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OrderWorktrees(wts, opt)
	}
}

func BenchmarkOrderWorktrees_ByActivity(b *testing.B) {
	wts := benchWorktrees(100)
	now := time.Now()
	activity := make(map[string]time.Time, len(wts))
	for i, wt := range wts {
		activity[wt.Path] = now.Add(-time.Duration(i) * time.Minute)
	}
	opt := OrderWorktreesOptions{
		Sort:     SortWorktreesByActivity,
		NormPath: func(p string) string { return p },
		ActivityAt: func(p string) time.Time {
			return activity[p]
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OrderWorktrees(wts, opt)
	}
}

func benchWorktrees(n int) []Worktree {
	out := make([]Worktree, n)
	for i := range out {
		name := fmt.Sprintf("wt-%03d", i)
		out[i] = Worktree{
			Path:    "/bench/" + name,
			DirName: name,
			Main:    i == 0,
		}
	}
	return out
}