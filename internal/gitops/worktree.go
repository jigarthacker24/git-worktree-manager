package gitops

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Worktree struct {
	Path    string
	DirName string
	Branch  string
	Commit  string
	Main    bool
}

func NormalizePath(path string) (string, error) {
	return filepath.Abs(filepath.Clean(path))
}

func IsRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir() || info.Mode()&os.ModeSymlink != 0
}

func ListWorktrees(repoPath string) ([]Worktree, error) {
	out, err := runGit(repoPath, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktrees(out), nil
}

func ListBranches(repoPath string) ([]string, error) {
	out, err := runGit(repoPath, "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

func AddWorktree(repoPath, wtPath, branch string, newBranch bool, sourceBranch string) error {
	args := addWorktreeArgs(wtPath, branch, newBranch, sourceBranch)
	_, err := runGit(repoPath, args...)
	return err
}

func addWorktreeArgs(wtPath, branch string, newBranch bool, sourceBranch string) []string {
	args := []string{"worktree", "add"}
	if newBranch {
		args = append(args, "-b", branch, wtPath)
		if sourceBranch != "" {
			args = append(args, sourceBranch)
		}
		return args
	}
	return append(args, wtPath, branch)
}

func RemoveWorktree(repoPath, wtPath string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, wtPath)
	_, err := runGit(repoPath, args...)
	return err
}

// WorktreeMove describes relocating one linked worktree into the default directory.
type WorktreeMove struct {
	From string
	To   string
}

// PlanWorktreeMoves returns moves for linked worktrees not already direct children of baseDir.
// The main worktree is never moved.
func PlanWorktreeMoves(worktrees []Worktree, baseDir string) ([]WorktreeMove, error) {
	base, err := NormalizePath(baseDir)
	if err != nil {
		return nil, err
	}
	if base == "" {
		return nil, fmt.Errorf("worktree directory is required")
	}

	var moves []WorktreeMove
	seenDest := make(map[string]string)
	for _, wt := range worktrees {
		if wt.Main {
			continue
		}
		parent, err := NormalizePath(filepath.Dir(wt.Path))
		if err != nil {
			return nil, err
		}
		if parent == base {
			continue
		}
		dest, err := NormalizePath(filepath.Join(base, wt.DirName))
		if err != nil {
			return nil, err
		}
		if dest == wt.Path {
			continue
		}
		if other, ok := seenDest[dest]; ok {
			return nil, fmt.Errorf("multiple worktrees would move to %s (%s and %s)", dest, other, wt.Path)
		}
		seenDest[dest] = wt.Path
		moves = append(moves, WorktreeMove{From: wt.Path, To: dest})
	}
	return moves, nil
}

func MoveWorktree(repoPath, from, to string, force bool) error {
	_, err := runGit(repoPath, moveWorktreeArgs(from, to, force)...)
	return err
}

func moveWorktreeArgs(from, to string, force bool) []string {
	args := []string{"worktree", "move"}
	if force {
		args = append(args, "--force")
	}
	return append(args, from, to)
}

func runGit(repoPath string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}
	return stdout.String(), nil
}

func parseWorktrees(out string) []Worktree {
	var worktrees []Worktree
	var current Worktree

	flush := func() {
		if current.Path != "" {
			worktrees = append(worktrees, current)
		}
		current = Worktree{}
	}

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			flush()
			continue
		}
		switch {
		case strings.HasPrefix(line, "worktree "):
			current.Path = strings.TrimPrefix(line, "worktree ")
			if norm, err := NormalizePath(current.Path); err == nil {
				current.Path = norm
			}
			current.DirName = filepath.Base(current.Path)
		case strings.HasPrefix(line, "HEAD "):
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "bare":
			current.Main = true
		}
	}
	flush()

	if len(worktrees) > 0 {
		worktrees[0].Main = true
	}
	return worktrees
}
