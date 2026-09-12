package ide

import "testing"

func TestClaudeDeepLink(t *testing.T) {
	got := claudeDeepLink("/repo/my worktree")
	want := "claude-cli://open?cwd=%2Frepo%2Fmy+worktree"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestClaudeDesktopDeepLink(t *testing.T) {
	got := claudeDesktopDeepLink("/repo/my worktree")
	want := "claude://code/new?folder=/repo/my+worktree"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDoubleClickIDEPrefs(t *testing.T) {
	tests := []struct {
		pref  string
		label string
		ok    bool
		kind  Kind
	}{
		{pref: "", label: "None", ok: false},
		{pref: "vscode", label: "VS Code", ok: true, kind: VSCode},
		{pref: "cursor", label: "Cursor", ok: true, kind: Cursor},
		{pref: "claude", label: "Claude Code", ok: true, kind: Claude},
	}
	for _, tc := range tests {
		if got := LabelFromPref(tc.pref); got != tc.label {
			t.Fatalf("LabelFromPref(%q) = %q, want %q", tc.pref, got, tc.label)
		}
		if got := PrefFromLabel(tc.label); got != tc.pref {
			t.Fatalf("PrefFromLabel(%q) = %q, want %q", tc.label, got, tc.pref)
		}
		kind, ok := KindFromPref(tc.pref)
		if ok != tc.ok || (tc.ok && kind != tc.kind) {
			t.Fatalf("KindFromPref(%q) = (%v, %v), want (%v, %v)", tc.pref, kind, ok, tc.kind, tc.ok)
		}
	}
}

func TestEncodeClaudeFolder(t *testing.T) {
	got := encodeClaudeFolder("/Users/me/project")
	want := "/Users/me/project"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
