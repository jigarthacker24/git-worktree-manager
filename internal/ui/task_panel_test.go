package ui

import (
	"testing"

	"fyne.io/fyne/v2/widget"
)

func TestStrikeThrough(t *testing.T) {
	got := strikeThrough("abc")
	if got == "abc" {
		t.Fatal("expected strike-through combining marks")
	}
	if len([]rune(got)) != 6 {
		t.Fatalf("expected 6 runes, got %d (%q)", len([]rune(got)), got)
	}
}

func TestApplyTaskNameLabelActive(t *testing.T) {
	lbl := widget.NewLabel("")
	applyTaskNameLabel(lbl, "Active task", false)
	if lbl.Text != "Active task" {
		t.Fatalf("unexpected text: %q", lbl.Text)
	}
	if lbl.TextStyle.Italic {
		t.Fatal("active task should not be italic")
	}
}

func TestApplyTaskNameLabelCompleted(t *testing.T) {
	lbl := widget.NewLabel("")
	applyTaskNameLabel(lbl, "Done task", true)
	if lbl.Text == "Done task" {
		t.Fatal("expected struck-through text")
	}
	if !lbl.TextStyle.Italic {
		t.Fatal("completed task should be italic")
	}
}

func TestFmtTaskCount(t *testing.T) {
	if fmtTaskCount(1) != "1 task" {
		t.Fatal("singular")
	}
	if fmtTaskCount(3) != "3 tasks" {
		t.Fatal("plural")
	}
}

func TestStrikeThroughNotIdentity(t *testing.T) {
	if strikeThrough("abc") == "abc" {
		t.Fatal("expected modified text")
	}
}
