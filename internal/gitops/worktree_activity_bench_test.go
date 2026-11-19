package gitops

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestBenchRepoAllWorktrees times parallel last-activity for every linked worktree (one shot):
//
//	BENCH_REPO_ROOT=~/ng-src/ng-backend go test -run=TestBenchRepoAllWorktrees -v ./internal/gitops/
func TestBenchRepoAllWorktrees(t *testing.T) {
	root := os.Getenv("BENCH_REPO_ROOT")
	if root == "" {
		t.Skip("set BENCH_REPO_ROOT")
	}
	wts, err := ListWorktrees(root)
	if err != nil {
		t.Fatal(err)
	}
	paths := PathsNeedingActivity(wts, nil, filepath.Clean)
	t.Logf("worktrees: %d (activity checks for %d, excluding main)", len(wts), len(paths))
	start := time.Now()
	m := ComputeWorktreeLastActivity(paths)
	elapsed := time.Since(start)
	t.Logf("ComputeWorktreeLastActivity elapsed: %s (avg %s per worktree)", elapsed, elapsed/time.Duration(len(paths)))
	for _, wt := range wts {
		at := m[filepath.Clean(wt.Path)]
		t.Logf("  %s -> %s", wt.DirName, at.Format(time.RFC3339))
	}
}

// BenchmarkWorktreeLastActivity on your machine against a real worktree:
//
//	BENCH_WORKTREE_PATH=/path/to/wt go test -bench=BenchmarkWorktreeLastActivity_Real -benchmem ./internal/gitops/
func BenchmarkWorktreeLastActivity_Real(b *testing.B) {
	wt := os.Getenv("BENCH_WORKTREE_PATH")
	if wt == "" {
		b.Skip("set BENCH_WORKTREE_PATH")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WorktreeLastActivity(wt)
	}
}

func BenchmarkWorktreeLastActivity_CleanMiniRepo(b *testing.B) {
	root := makeBenchRepo(b, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WorktreeLastActivity(root)
	}
}

func BenchmarkWorktreeLastActivity_DirtyMiniRepo(b *testing.B) {
	root := makeBenchRepo(b, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WorktreeLastActivity(root)
	}
}

func BenchmarkComputeWorktreeLastActivity(b *testing.B) {
	sizes := []int{1, 8, 21, 100}
	for _, n := range sizes {
		n := n
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			paths := make([]string, n)
			for i := range paths {
				paths[i] = makeBenchRepo(b, i%3 == 0) // some dirty
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ComputeWorktreeLastActivity(paths)
			}
		})
	}
}

func makeBenchRepo(b *testing.B, dirty bool) string {
	b.Helper()
	root := b.TempDir()
	emptyTemplate := filepath.Join(root, "_template")
	if err := os.Mkdir(emptyTemplate, 0o755); err != nil {
		b.Fatal(err)
	}
	mustGit(b, root, []string{"GIT_TEMPLATE_DIR=" + emptyTemplate}, "init", "-b", "main")
	mustGit(b, root, nil, "config", "user.email", "bench@local")
	mustGit(b, root, nil, "config", "user.name", "bench")
	file := filepath.Join(root, "README.md")
	if err := os.WriteFile(file, []byte("bench"), 0o644); err != nil {
		b.Fatal(err)
	}
	mustGit(b, root, nil, "add", "README.md")
	mustGit(b, root, nil, "commit", "-m", "init")
	if dirty {
		if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("bench dirty\n"), 0o644); err != nil {
			b.Fatal(err)
		}
	}
	return root
}

func mustGit(b *testing.B, dir string, extraEnv []string, args ...string) {
	b.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		b.Fatalf("git %v: %s", args, msg)
	}
}