package gitops

import (
	"sort"
	"strings"
	"time"
)

// WorktreeListSort controls ordering of non-main, unpinned worktrees.
type WorktreeListSort string

const (
	SortWorktreesByName     WorktreeListSort = "name"
	SortWorktreesByActivity WorktreeListSort = "activity"
)

const (
	sortLabelByName     = "Directory name"
	sortLabelByActivity = "Last activity"
)

// WorktreeListSortLabels are settings UI options.
func WorktreeListSortLabels() []string {
	return []string{sortLabelByName, sortLabelByActivity}
}

// WorktreeListSortLabel returns the settings label for a sort mode.
func WorktreeListSortLabel(s WorktreeListSort) string {
	if s == SortWorktreesByActivity {
		return sortLabelByActivity
	}
	return sortLabelByName
}

// WorktreeListSortFromLabel parses a settings label.
func WorktreeListSortFromLabel(label string) WorktreeListSort {
	if strings.TrimSpace(label) == sortLabelByActivity {
		return SortWorktreesByActivity
	}
	return SortWorktreesByName
}

// ParseWorktreeListSort returns sort mode from preference value (default: by name).
func ParseWorktreeListSort(v string) WorktreeListSort {
	switch WorktreeListSort(strings.TrimSpace(v)) {
	case SortWorktreesByActivity:
		return SortWorktreesByActivity
	default:
		return SortWorktreesByName
	}
}

// OrderWorktreesOptions configures worktree list ordering.
type OrderWorktreesOptions struct {
	Sort       WorktreeListSort
	Pinned     []string // normalized paths, pin order
	NormPath   func(string) string
	ActivityAt func(path string) time.Time
}

// OrderWorktrees returns worktrees with main first, then pinned (excluding main), then the rest sorted.
func OrderWorktrees(wts []Worktree, opt OrderWorktreesOptions) []Worktree {
	if opt.NormPath == nil {
		opt.NormPath = func(p string) string { return p }
	}
	if opt.ActivityAt == nil {
		opt.ActivityAt = func(string) time.Time { return time.Time{} }
	}

	byPath := make(map[string]Worktree, len(wts))
	var main *Worktree
	for _, wt := range wts {
		key := opt.NormPath(wt.Path)
		byPath[key] = wt
		if wt.Main {
			copy := wt
			main = &copy
		}
	}

	pinOrder := make(map[string]int, len(opt.Pinned))
	for i, p := range opt.Pinned {
		pinOrder[opt.NormPath(p)] = i
	}

	var out []Worktree
	if main != nil {
		out = append(out, *main)
		mainKey := opt.NormPath(main.Path)
		delete(byPath, mainKey)
	}

	for _, p := range opt.Pinned {
		key := opt.NormPath(p)
		if main != nil && key == opt.NormPath(main.Path) {
			continue
		}
		if wt, ok := byPath[key]; ok {
			out = append(out, wt)
			delete(byPath, key)
		}
	}

	rest := make([]Worktree, 0, len(byPath))
	for _, wt := range byPath {
		rest = append(rest, wt)
	}

	switch opt.Sort {
	case SortWorktreesByActivity:
		sort.Slice(rest, func(i, j int) bool {
			ti := opt.ActivityAt(rest[i].Path)
			tj := opt.ActivityAt(rest[j].Path)
			if !ti.Equal(tj) {
				return ti.After(tj)
			}
			return strings.ToLower(rest[i].DirName) < strings.ToLower(rest[j].DirName)
		})
	default:
		sort.Slice(rest, func(i, j int) bool {
			return strings.ToLower(rest[i].DirName) < strings.ToLower(rest[j].DirName)
		})
	}
	out = append(out, rest...)
	return out
}