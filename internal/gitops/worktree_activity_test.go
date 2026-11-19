package gitops

import "testing"

func TestPathsFromPorcelainZ(t *testing.T) {
	out := " M src/a.go\x00?? notes.txt\x00R  old.txt\x00new.txt\x00"
	got := pathsFromPorcelainZ(out)
	want := []string{"src/a.go", "notes.txt", "old.txt", "new.txt"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestComputeWorktreeLastActivityEmpty(t *testing.T) {
	m := ComputeWorktreeLastActivity(nil)
	if len(m) != 0 {
		t.Fatalf("got %v", m)
	}
}

func TestPathsFromPorcelainZDedupe(t *testing.T) {
	out := "M  same\x00 M same\x00"
	got := pathsFromPorcelainZ(out)
	if len(got) != 1 || got[0] != "same" {
		t.Fatalf("got %v", got)
	}
}