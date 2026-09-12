package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// DoubleTapZone wraps content and calls OnDoubleTapped on desktop double-click.
type DoubleTapZone struct {
	widget.BaseWidget
	content        fyne.CanvasObject
	OnDoubleTapped func()
}

// NewDoubleTapZone wraps content for double-click handling.
func NewDoubleTapZone(content fyne.CanvasObject) *DoubleTapZone {
	z := &DoubleTapZone{content: content}
	z.ExtendBaseWidget(z)
	return z
}

// Content returns the wrapped object.
func (z *DoubleTapZone) Content() fyne.CanvasObject {
	return z.content
}

func (z *DoubleTapZone) CreateRenderer() fyne.WidgetRenderer {
	return &doubleTapRenderer{zone: z}
}

func (z *DoubleTapZone) DoubleTapped(*fyne.PointEvent) {
	if z.OnDoubleTapped != nil {
		z.OnDoubleTapped()
	}
}

// ListRowCenter returns the worktree column container from a list row border.
func ListRowCenter(border *fyne.Container) *fyne.Container {
	if border == nil || len(border.Objects) == 0 {
		return nil
	}
	center := border.Objects[0]
	if zone, ok := center.(*DoubleTapZone); ok {
		if cols, ok := zone.Content().(*fyne.Container); ok {
			return cols
		}
		return nil
	}
	if cols, ok := center.(*fyne.Container); ok {
		return cols
	}
	return nil
}

type doubleTapRenderer struct {
	zone *DoubleTapZone
}

func (r *doubleTapRenderer) Layout(size fyne.Size) {
	r.zone.content.Resize(size)
	r.zone.content.Move(fyne.NewPos(0, 0))
}

func (r *doubleTapRenderer) MinSize() fyne.Size {
	return r.zone.content.MinSize()
}

func (r *doubleTapRenderer) Refresh() {
	r.zone.content.Refresh()
}

func (r *doubleTapRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.zone.content}
}

func (r *doubleTapRenderer) Destroy() {}

var _ fyne.DoubleTappable = (*DoubleTapZone)(nil)
