package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ReadOnlyPathField looks like an entry but cannot be edited and keeps readable text color.
type ReadOnlyPathField struct {
	widget.BaseWidget
	label *widget.Label
}

func NewReadOnlyPathField(text string) *ReadOnlyPathField {
	f := &ReadOnlyPathField{}
	f.label = widget.NewLabel(text)
	f.label.Truncation = fyne.TextTruncateEllipsis
	f.ExtendBaseWidget(f)
	return f
}

func (f *ReadOnlyPathField) SetText(text string) {
	f.label.SetText(text)
}

func (f *ReadOnlyPathField) Text() string {
	return f.label.Text
}

// SetTruncate controls whether long paths are shortened with an ellipsis.
func (f *ReadOnlyPathField) SetTruncate(truncate bool) {
	if truncate {
		f.label.Truncation = fyne.TextTruncateEllipsis
	} else {
		f.label.Truncation = fyne.TextTruncateOff
	}
	f.Refresh()
}

func (f *ReadOnlyPathField) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.StrokeColor = theme.Color(theme.ColorNameInputBorder)
	bg.StrokeWidth = theme.InputBorderSize()
	bg.CornerRadius = theme.InputRadiusSize()
	return &readOnlyPathRenderer{field: f, bg: bg, label: f.label}
}

type readOnlyPathRenderer struct {
	field   *ReadOnlyPathField
	bg      *canvas.Rectangle
	label   *widget.Label
	objects []fyne.CanvasObject
}

func (r *readOnlyPathRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.bg.Move(fyne.NewPos(0, 0))
	pad := theme.Padding()
	r.label.Resize(fyne.NewSize(size.Width-pad*2, size.Height))
	r.label.Move(fyne.NewPos(pad, (size.Height-r.label.MinSize().Height)/2))
}

func (r *readOnlyPathRenderer) MinSize() fyne.Size {
	pad := theme.Padding()
	labelMin := r.label.MinSize()
	w := labelMin.Width + pad*2
	h := widget.NewEntry().MinSize().Height
	return fyne.NewSize(w, h)
}

func (r *readOnlyPathRenderer) Refresh() {
	r.bg.FillColor = theme.Color(theme.ColorNameInputBackground)
	r.bg.StrokeColor = theme.Color(theme.ColorNameInputBorder)
	r.bg.StrokeWidth = theme.InputBorderSize()
	r.bg.CornerRadius = theme.InputRadiusSize()
	r.bg.Refresh()
	r.label.Refresh()
}

func (r *readOnlyPathRenderer) Objects() []fyne.CanvasObject {
	if r.objects == nil {
		r.objects = []fyne.CanvasObject{r.bg, r.label}
	}
	return r.objects
}

func (r *readOnlyPathRenderer) Destroy() {}
