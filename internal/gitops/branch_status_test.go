package gitops

import (
	"strings"
	"testing"
)

func TestSubjectPreview(t *testing.T) {
	got := SubjectPreview("fix the login bug for oauth users", 3)
	if got != "fix the login…" {
		t.Fatalf("got %q", got)
	}
	if SubjectPreview("one", 5) != "one" {
		t.Fatal("short subject unchanged")
	}
}

func TestBranchSourceStatusSyncLine(t *testing.T) {
	up := BranchSourceStatus{HasRemote: true, RemoteRef: "origin/main", BehindRemote: 0, AheadRemote: 0}
	if up.SyncLine() != "Up to date with origin/main" {
		t.Fatal(up.SyncLine())
	}
	behind := BranchSourceStatus{HasRemote: true, RemoteRef: "origin/develop", BehindRemote: 3}
	if behind.SyncLine() != "3 commit(s) behind origin/develop" {
		t.Fatal(behind.SyncLine())
	}
}

func TestBranchSourceStatusSummary(t *testing.T) {
	s := BranchSourceStatus{
		HasRemote:      true,
		RemoteRef:      "origin/main",
		ShortSHA:       "abc1234",
		Author:         "Ada",
		SubjectPreview: "fix login flow",
	}
	sum := s.Summary()
	if !strings.Contains(sum, "Up to date") || !strings.Contains(sum, "abc1234") {
		t.Fatalf("summary: %s", sum)
	}
}
