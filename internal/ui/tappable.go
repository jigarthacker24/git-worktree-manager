package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var hoverBorderColor = color.NRGBA{R: 255, G: 204, B: 0, A: 255}

// TappableRow is a two-column row that opens on click and shows a yellow border on hover.
type TappableRow struct {
	widget.BaseWidget
	dirName  string
	path     string
	OnTapped func()
	hovered  bool
}

// NewTappableRow returns a row with dir name and path columns.
func NewTappableRow(dirName, path string, onTapped func()) fyne.CanvasObject {
	r := &TappableRow{dirName: dirName, path: path, OnTapped: onTapped}
	r.ExtendBaseWidget(r)
	return r
}

func (r *TappableRow) CreateRenderer() fyne.WidgetRenderer {
	border := canvas.NewRectangle(color.Transparent)
	border.StrokeWidth = 0
	dirText := canvas.NewText(r.dirName, theme.ForegroundColor())
	dirText.TextSize = theme.TextSize()
	pathText := canvas.NewText(r.path, theme.ForegroundColor())
	pathText.TextSize = theme.TextSize()
	return &tappableRowRenderer{
		row:      r,
		border:   border,
		dirText:  dirText,
		pathText: pathText,
	}
}

func (r *TappableRow) Tapped(*fyne.PointEvent) {
	if r.OnTapped != nil {
		r.OnTapped()
	}
}

func (r *TappableRow) MouseIn(*desktop.MouseEvent) {
	if r.hovered {
		return
	}
	r.hovered = true
	r.Refresh()
}

func (r *TappableRow) MouseOut() {
	if !r.hovered {
		return
	}
	r.hovered = false
	r.Refresh()
}

func (r *TappableRow) MouseMoved(*desktop.MouseEvent) {}

type tappableRowRenderer struct {
	row      *TappableRow
	border   *canvas.Rectangle
	dirText  *canvas.Text
	pathText *canvas.Text
}

func (rr *tappableRowRenderer) Layout(size fyne.Size) {
	const borderInset float32 = 1

	rr.border.Resize(size.Subtract(fyne.NewSize(borderInset*2, borderInset*2)))
	rr.border.Move(fyne.NewPos(borderInset, borderInset))

	pad := theme.Padding()
	innerW := size.Width - pad*2
	innerH := size.Height - pad*2
	colW := innerW / 2
	textH := rr.dirText.MinSize().Height

	rr.dirText.Text = truncateText(rr.row.dirName, colW, rr.dirText.TextSize)
	rr.pathText.Text = truncateText(rr.row.path, colW, rr.pathText.TextSize)

	rr.dirText.Resize(fyne.NewSize(colW, textH))
	rr.dirText.Move(fyne.NewPos(pad, pad+(innerH-textH)/2))

	rr.pathText.Resize(fyne.NewSize(colW, textH))
	rr.pathText.Move(fyne.NewPos(pad+colW, pad+(innerH-textH)/2))
}

func (rr *tappableRowRenderer) MinSize() fyne.Size {
	pad := theme.Padding()
	textH := rr.dirText.MinSize().Height
	return fyne.NewSize(200, textH+pad*2)
}

func (rr *tappableRowRenderer) Refresh() {
	if rr.row.hovered {
		rr.border.StrokeColor = hoverBorderColor
		rr.border.StrokeWidth = 2
	} else {
		rr.border.StrokeColor = color.Transparent
		rr.border.StrokeWidth = 0
	}
	rr.border.Refresh()
	rr.dirText.Color = theme.ForegroundColor()
	rr.pathText.Color = theme.ForegroundColor()
	rr.dirText.Refresh()
	rr.pathText.Refresh()
}

func (rr *tappableRowRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{rr.dirText, rr.pathText, rr.border}
}

func (rr *tappableRowRenderer) Destroy() {}

func truncateText(text string, maxWidth, textSize float32) string {
	if maxWidth <= 0 {
		return text
	}
	charW := textSize * 0.55
	maxChars := int(maxWidth / charW)
	if maxChars < 1 {
		return text
	}
	runes := []rune(text)
	if len(runes) <= maxChars {
		return text
	}
	if maxChars <= 3 {
		return string(runes[:maxChars])
	}
	return string(runes[:maxChars-3]) + "..."
}
