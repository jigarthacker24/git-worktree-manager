package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

type expandHBoxLayout struct{}

// NewExpandHBox lays out objects in a row; the first object receives all remaining width.
func NewExpandHBox(objects ...fyne.CanvasObject) *fyne.Container {
	return container.New(&expandHBoxLayout{}, objects...)
}

func (l *expandHBoxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	gap := theme.Padding()
	fixedW := float32(0)
	for i := 1; i < len(objects); i++ {
		fixedW += objects[i].MinSize().Width
		if i < len(objects)-1 {
			fixedW += gap
		}
	}
	if len(objects) > 1 {
		fixedW += gap
	}
	firstW := size.Width - fixedW
	if firstW < 0 {
		firstW = 0
	}
	x := float32(0)
	objects[0].Resize(fyne.NewSize(firstW, size.Height))
	objects[0].Move(fyne.NewPos(x, 0))
	x += firstW
	for i := 1; i < len(objects); i++ {
		if i > 0 {
			x += gap
		}
		ms := objects[i].MinSize()
		objects[i].Resize(fyne.NewSize(ms.Width, size.Height))
		objects[i].Move(fyne.NewPos(x, 0))
		x += ms.Width
	}
}

func (l *expandHBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	gap := theme.Padding()
	width := objects[0].MinSize().Width
	height := objects[0].MinSize().Height
	for i := 1; i < len(objects); i++ {
		ms := objects[i].MinSize()
		width += gap + ms.Width
		if ms.Height > height {
			height = ms.Height
		}
	}
	return fyne.NewSize(width, height)
}

// NewRatioRow places two columns side by side with the given width ratio for the first column.
func NewRatioRow(firstRatio float32, first, second fyne.CanvasObject) *fyne.Container {
	return container.New(&ratioRowLayout{firstRatio: firstRatio}, first, second)
}

type ratioRowLayout struct {
	firstRatio float32
}

func (l *ratioRowLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	gap := theme.Padding()
	firstW := (size.Width - gap) * l.firstRatio
	secondW := size.Width - gap - firstW
	firstH := objects[0].MinSize().Height
	secondH := objects[1].MinSize().Height
	height := firstH
	if secondH > height {
		height = secondH
	}
	objects[0].Resize(fyne.NewSize(firstW, height))
	objects[0].Move(fyne.NewPos(0, 0))
	objects[1].Resize(fyne.NewSize(secondW, height))
	objects[1].Move(fyne.NewPos(firstW+gap, 0))
}

func (l *ratioRowLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.NewSize(0, 0)
	}
	gap := theme.Padding()
	first := objects[0].MinSize()
	second := objects[1].MinSize()
	height := first.Height
	if second.Height > height {
		height = second.Height
	}
	minTotalW := first.Width/l.firstRatio
	alt := second.Width / (1 - l.firstRatio)
	if alt > minTotalW {
		minTotalW = alt
	}
	return fyne.NewSize(minTotalW+gap, height)
}
