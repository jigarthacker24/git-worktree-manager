package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type hintWidget struct {
	widget.BaseWidget
	child  fyne.CanvasObject
	hint   string
	onHint func(string)
}

func WrapWithHint(child fyne.CanvasObject, hint string, onHint func(string)) fyne.CanvasObject {
	h := &hintWidget{child: child, hint: hint, onHint: onHint}
	h.ExtendBaseWidget(h)
	return h
}

func (h *hintWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.child)
}

func (h *hintWidget) MouseIn(*desktop.MouseEvent) {
	if h.onHint != nil {
		h.onHint(h.hint)
	}
}

func (h *hintWidget) Tapped(ev *fyne.PointEvent) {
	if t, ok := h.child.(fyne.Tappable); ok {
		t.Tapped(ev)
	}
}

func (h *hintWidget) TappedSecondary(ev *fyne.PointEvent) {
	if t, ok := h.child.(fyne.SecondaryTappable); ok {
		t.TappedSecondary(ev)
	}
}

func IconButton(icon fyne.Resource, hint string, onTap func(), onHint func(string)) fyne.CanvasObject {
	btn := widget.NewButtonWithIcon("", icon, onTap)
	btn.Importance = widget.LowImportance
	return WrapWithHint(btn, hint, onHint)
}

func (h *hintWidget) SetHint(hint string) {
	h.hint = hint
}

func SetHint(wrapped fyne.CanvasObject, hint string) {
	if h, ok := wrapped.(*hintWidget); ok {
		h.SetHint(hint)
	}
}

func ButtonFromHint(wrapped fyne.CanvasObject) *widget.Button {
	h, ok := wrapped.(*hintWidget)
	if !ok {
		return nil
	}
	btn, _ := h.child.(*widget.Button)
	return btn
}

func ButtonsFromHintHBox(hbox *fyne.Container) []*widget.Button {
	btns := make([]*widget.Button, 0, len(hbox.Objects))
	for _, obj := range hbox.Objects {
		if btn := ButtonFromHint(obj); btn != nil {
			btns = append(btns, btn)
		}
	}
	return btns
}
