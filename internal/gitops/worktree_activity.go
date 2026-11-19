package gitops

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxActivityWorkers = 8

// WorktreeLastActivity returns a best-effort "recent work" time for a worktree:
// the later of the latest HEAD commit time and the newest mtime among paths
// reported by git status (modified, staged, untracked).
func WorktreeLastActivity(wtPath string) time.Time {
	return worktreeLastActivity(wtPath)
}

// PathsNeedingActivity returns paths that are sorted by last activity (not main or pinned).
func PathsNeedingActivity(wts []Worktree, pinned []string, norm func(string) string) []string {
	if norm == nil {
		norm = func(p string) string { return filepath.Clean(p) }
	}
	pinSet := make(map[string]struct{}, len(pinned))
	for _, p := range pinned {
		pinSet[norm(p)] = struct{}{}
	}
	var paths []string
	for _, wt := range wts {
		if wt.Main {
			continue
		}
		if _, ok := pinSet[norm(wt.Path)]; ok {
			continue
		}
		paths = append(paths, wt.Path)
	}
	return paths
}

// ComputeWorktreeLastActivity computes last-activity times for many worktrees in parallel.
func ComputeWorktreeLastActivity(paths []string) map[string]time.Time {
	out := make(map[string]time.Time, len(paths))
	if len(paths) == 0 {
		return out
	}
	workers := maxActivityWorkers
	if n := runtime.GOMAXPROCS(0); n > 0 && n < workers {
		workers = n
	}
	if len(paths) < workers {
		workers = len(paths)
	}
	sem := make(chan struct{}, workers)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, p := range paths {
		path := filepath.Clean(p)
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			t := worktreeLastActivity(path)
			mu.Lock()
			out[path] = t
			mu.Unlock()
		}(path)
	}
	wg.Wait()
	return out
}

func worktreeLastActivity(wtPath string) time.Time {
	wtPath = filepath.Clean(wtPath)
	var (
		latest    time.Time
		paths     []string
		statusErr error
	)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		latest = lastCommitTime(wtPath)
	}()
	go func() {
		defer wg.Done()
		paths, statusErr = statusChangedPaths(wtPath)
	}()
	wg.Wait()
	if statusErr != nil {
		return latest
	}
	for _, rel := range paths {
		abs := filepath.Join(wtPath, rel)
		info, err := os.Stat(abs)
		if err != nil {
			continue
		}
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
	}
	return latest
}

func lastCommitTime(wtPath string) time.Time {
	out, err := runGit(wtPath, "log", "-1", "--format=%ct")
	if err != nil {
		return time.Time{}
	}
	secs, err := strconv.ParseInt(strings.TrimSpace(out), 10, 64)
	if err != nil || secs <= 0 {
		return time.Time{}
	}
	return time.Unix(secs, 0)
}

func statusChangedPaths(wtPath string) ([]string, error) {
	// Normal untracked mode (not -uall) keeps status fast on large worktrees.
	out, err := runGit(wtPath, "status", "--porcelain=1", "-z")
	if err != nil {
		return nil, err
	}
	return pathsFromPorcelainZ(out), nil
}

// pathsFromPorcelainZ extracts paths from git status --porcelain=1 -z output.
func pathsFromPorcelainZ(out string) []string {
	var paths []string
	seen := make(map[string]struct{})
	add := func(p string) {
		p = filepath.ToSlash(strings.TrimSpace(p))
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		paths = append(paths, p)
	}

	chunks := strings.Split(out, "\x00")
	for i := 0; i < len(chunks); i++ {
		chunk := chunks[i]
		if chunk == "" {
			continue
		}
		if len(chunk) < 4 || chunk[2] != ' ' {
			continue
		}
		xy := chunk[:2]
		path := chunk[3:]
		isRename := xy[0] == 'R' || xy[0] == 'C' || xy[1] == 'R' || xy[1] == 'C'
		if isRename {
			add(path)
			if i+1 < len(chunks) && chunks[i+1] != "" {
				add(chunks[i+1])
				i++
			}
			continue
		}
		add(path)
	}
	return paths
}