package gitops

import (
	"fmt"
	"strconv"
	"strings"
)

const subjectPreviewWords = 5

// BranchSourceStatus describes a source branch relative to its remote and latest commit.
type BranchSourceStatus struct {
	RemoteRef      string
	BehindRemote   int
	AheadRemote    int
	HasRemote      bool
	ShortSHA       string
	Author         string
	SubjectPreview string
}

// Summary returns multi-line text for the add-worktree dialog.
func (s BranchSourceStatus) Summary() string {
	var lines []string
	lines = append(lines, s.SyncLine())
	if s.ShortSHA != "" {
		lines = append(lines, s.CommitLine())
	}
	return strings.Join(lines, "\n")
}

// SyncLine describes remote sync state.
func (s BranchSourceStatus) SyncLine() string {
	if !s.HasRemote {
		return "No remote tracking branch found"
	}
	if s.RemoteRef != "" && s.BehindRemote == 0 && s.AheadRemote == 0 {
		return fmt.Sprintf("Up to date with %s", s.RemoteRef)
	}
	if s.BehindRemote > 0 && s.AheadRemote == 0 {
		return fmt.Sprintf("%d commit(s) behind %s", s.BehindRemote, s.RemoteRef)
	}
	if s.BehindRemote == 0 && s.AheadRemote > 0 {
		return fmt.Sprintf("%d commit(s) ahead of %s", s.AheadRemote, s.RemoteRef)
	}
	if s.BehindRemote > 0 && s.AheadRemote > 0 {
		return fmt.Sprintf("%d behind, %d ahead of %s", s.BehindRemote, s.AheadRemote, s.RemoteRef)
	}
	if s.RemoteRef != "" {
		return fmt.Sprintf("Tracking %s", s.RemoteRef)
	}
	return ""
}

// CommitLine formats the latest commit on the source ref.
func (s BranchSourceStatus) CommitLine() string {
	preview := s.SubjectPreview
	if preview == "" {
		preview = "(no message)"
	}
	return fmt.Sprintf("Latest: %s · %s · %s", s.ShortSHA, s.Author, preview)
}

// DescribeBranchSource loads sync and commit info for a checkout branch option.
func DescribeBranchSource(repoPath string, opt BranchOption) (BranchSourceStatus, error) {
	ref := opt.SourceRefForNewBranch()
	if ref == "" {
		return BranchSourceStatus{}, fmt.Errorf("branch ref is required")
	}

	status := BranchSourceStatus{}
	if err := fillLatestCommit(repoPath, ref, &status); err != nil {
		return status, err
	}
	fillRemoteSync(repoPath, opt, &status)
	return status, nil
}

func fillLatestCommit(repoPath, ref string, status *BranchSourceStatus) error {
	out, err := runGit(repoPath, "log", "-1", "--format=%h|%an|%s", ref)
	if err != nil {
		return err
	}
	parts := strings.SplitN(strings.TrimSpace(out), "|", 3)
	if len(parts) < 3 {
		return fmt.Errorf("unexpected git log output")
	}
	status.ShortSHA = parts[0]
	status.Author = parts[1]
	status.SubjectPreview = SubjectPreview(parts[2], subjectPreviewWords)
	return nil
}

func fillRemoteSync(repoPath string, opt BranchOption, status *BranchSourceStatus) {
	if opt.RemoteOnly {
		status.HasRemote = true
		status.RemoteRef = opt.Ref
		status.BehindRemote = 0
		status.AheadRemote = 0
		return
	}

	localRef := opt.Ref
	remoteRef, ok := resolveUpstreamRef(repoPath, localRef, opt.LocalName)
	if !ok {
		return
	}
	status.HasRemote = true
	status.RemoteRef = remoteRef
	status.BehindRemote = revListCount(repoPath, localRef+".."+remoteRef)
	status.AheadRemote = revListCount(repoPath, remoteRef+".."+localRef)
}

func resolveUpstreamRef(repoPath, localRef, localName string) (string, bool) {
	out, err := runGit(repoPath, "rev-parse", "--abbrev-ref", "--symbolic-full-name", localRef+"@{upstream}")
	if err == nil {
		ref := strings.TrimSpace(out)
		if ref != "" && !strings.Contains(ref, "fatal") {
			return ref, true
		}
	}
	candidate := "origin/" + localName
	if _, err := runGit(repoPath, "rev-parse", "--verify", candidate); err == nil {
		return candidate, true
	}
	return "", false
}

func revListCount(repoPath, rangeExpr string) int {
	out, err := runGit(repoPath, "rev-list", "--count", rangeExpr)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// SubjectPreview returns the first maxWords words of a commit subject.
func SubjectPreview(subject string, maxWords int) string {
	if maxWords < 1 {
		maxWords = 1
	}
	words := strings.Fields(strings.TrimSpace(subject))
	if len(words) == 0 {
		return ""
	}
	if len(words) <= maxWords {
		return strings.Join(words, " ")
	}
	return strings.Join(words[:maxWords], " ") + "…"
}
