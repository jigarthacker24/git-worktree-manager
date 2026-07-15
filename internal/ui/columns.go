package ui

import (
	"github.com/jigarthacker24/git-worktree-manager/internal/gitops"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// RowMetrics holds measured widths for the worktree list row.
type RowMetrics struct {
	DirWidth     float32
	TruncatePath bool
}

// ColumnGap is the space between columns in the center row.
func ColumnGap() float32 {
	return float32(theme.Padding()) * 2
}

type worktreeCenterLayout struct {
	dirWidth *float32
}

// NewWorktreeCenter lays out Dir (fixed) and Branch/Path (equal share of remaining width).
// Objects must be: dir, branch cell, path cell.
func NewWorktreeCenter(dirWidth *float32, objects ...fyne.CanvasObject) *fyne.Container {
	return container.New(&worktreeCenterLayout{dirWidth: dirWidth}, objects...)
}

func (l *worktreeCenterLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	gap := ColumnGap()
	dirW := float32(0)
	if l.dirWidth != nil {
		dirW = *l.dirWidth
	}

	x := float32(0)
	if len(objects) > 0 {
		objects[0].Resize(fyne.NewSize(dirW, size.Height))
		objects[0].Move(fyne.NewPos(x, 0))
		x += dirW + gap
	}

	flexW := size.Width - dirW - gap
	if len(objects) > 1 {
		flexW -= gap
	}
	if flexW < 0 {
		flexW = 0
	}
	half := flexW / 2

	if len(objects) > 1 {
		objects[1].Resize(fyne.NewSize(half, size.Height))
		objects[1].Move(fyne.NewPos(x, 0))
		x += half + gap
	}
	if len(objects) > 2 {
		objects[2].Resize(fyne.NewSize(flexW-half, size.Height))
		objects[2].Move(fyne.NewPos(x, 0))
	}
}

func (l *worktreeCenterLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	gap := ColumnGap()
	dirW := float32(0)
	if l.dirWidth != nil {
		dirW = *l.dirWidth
	}
	var maxH float32
	for _, obj := range objects {
		if h := obj.MinSize().Height; h > maxH {
			maxH = h
		}
	}
	minFlex := textWidth("Branch", true)
	if w := textWidth("Path", true); w > minFlex {
		minFlex = w
	}
	return fyne.NewSize(dirW+gap+minFlex+gap+minFlex, maxH)
}

// WorktreeBranchLabel returns the branch text shown in the worktree list.
func WorktreeBranchLabel(wt gitops.Worktree) string {
	branch := wt.Branch
	if branch == "" {
		branch = "(detached)"
	}
	if wt.Main {
		branch += " · main"
	}
	return branch
}

// ComputeDirWidth returns the width needed for the Dir column.
func ComputeDirWidth(wts []gitops.Worktree) float32 {
	pad := float32(theme.Padding())
	w := textWidth("Dir", true) + pad
	for _, wt := range wts {
		if tw := textWidth(wt.DirName, false) + pad; tw > w {
			w = tw
		}
	}
	return w
}

// ComputeTruncatePath reports whether any path needs ellipsis in the given flex half-width.
func ComputeTruncatePath(wts []gitops.Worktree, pathHalfWidth float32) bool {
	copyW := CopyColumnWidth()
	maxText := pathHalfWidth - copyW - float32(theme.Padding())
	if maxText <= 0 {
		return len(wts) > 0
	}
	for _, wt := range wts {
		if textWidth(wt.Path, false) > maxText {
			return true
		}
	}
	return false
}

// FlexHalfWidth returns the width of one flex column (branch or path) in the center area.
func FlexHalfWidth(centerWidth, dirWidth float32) float32 {
	flexW := centerWidth - dirWidth - ColumnGap()*2
	if flexW < 0 {
		return 0
	}
	return flexW / 2
}

// OpenColumnWidth estimates the width of the three IDE open buttons.
func OpenColumnWidth() float32 {
	btn := CopyColumnWidth()
	return btn*3 + float32(theme.Padding())*2
}

// NewTextWithCopy is a cell with a label and a copy button reserved on the right.
func NewTextWithCopy(label *widget.Label, copyBtn *widget.Button) *fyne.Container {
	label.Truncation = fyne.TextTruncateOff
	return container.NewBorder(nil, nil, nil, copyBtn, label)
}

// LabelFromTextCell returns the label inside a NewTextWithCopy container.
func LabelFromTextCell(cell *fyne.Container) *widget.Label {
	if len(cell.Objects) == 0 {
		return nil
	}
	lbl, _ := cell.Objects[0].(*widget.Label)
	return lbl
}

// ButtonFromTextCell returns the copy button inside a NewTextWithCopy container.
func ButtonFromTextCell(cell *fyne.Container) *widget.Button {
	if len(cell.Objects) < 2 {
		return nil
	}
	btn, _ := cell.Objects[1].(*widget.Button)
	return btn
}

// PinColumnSpacer matches the pin button width so headers align with list rows.
func PinColumnSpacer() fyne.CanvasObject {
	r := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	r.SetMinSize(fyne.NewSize(PinButtonWidth(), 1))
	return r
}

// TextWidth measures rendered text width.
func TextWidth(text string, bold bool) float32 {
	return textWidth(text, bold)
}

func textWidth(text string, bold bool) float32 {
	style := fyne.TextStyle{Bold: bold}
	t := canvas.NewText(text, theme.Color(theme.ColorNameForeground))
	t.TextStyle = style
	t.TextSize = theme.TextSize()
	return t.MinSize().Width
}

// CopyColumnWidth is the space reserved for a copy button inside a cell.
func CopyColumnWidth() float32 {
	return theme.IconInlineSize() + float32(theme.Padding())*4
}

// PinButtonWidth is the width of the pin column.
func PinButtonWidth() float32 {
	return theme.IconInlineSize() + float32(theme.Padding())*2
}
