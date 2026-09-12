package ui

import (
	fynetooltip "github.com/dweymouth/fyne-tooltip"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// WindowContent wraps content with the fyne-tooltip layer required for tooltips.
func WindowContent(window fyne.Window, content fyne.CanvasObject) fyne.CanvasObject {
	return fynetooltip.AddWindowToolTipLayer(content, window.Canvas())
}

// NewToolButton creates a text button with a tooltip.
func NewToolButton(text, tooltip string, onTap func()) *ttwidget.Button {
	btn := ttwidget.NewButton(text, onTap)
	btn.SetToolTip(tooltip)
	return btn
}

// NewToolButtonWithIcon creates a button with text, icon, and tooltip.
func NewToolButtonWithIcon(text string, icon fyne.Resource, tooltip string, onTap func()) *ttwidget.Button {
	btn := ttwidget.NewButtonWithIcon(text, icon, onTap)
	btn.SetToolTip(tooltip)
	return btn
}

// NewIconToolButton creates an icon-only button with a tooltip.
func NewIconToolButton(icon fyne.Resource, tooltip string, onTap func()) *ttwidget.Button {
	btn := ttwidget.NewButtonWithIcon("", icon, onTap)
	btn.Importance = widget.LowImportance
	btn.SetToolTip(tooltip)
	return btn
}

// SetToolTip updates a tooltip-enabled button's tooltip text.
func SetToolTip(btn *ttwidget.Button, tooltip string) {
	if btn != nil {
		btn.SetToolTip(tooltip)
	}
}

// ToolButtonFrom returns a tooltip button from a canvas object.
func ToolButtonFrom(obj fyne.CanvasObject) *ttwidget.Button {
	btn, _ := obj.(*ttwidget.Button)
	return btn
}

// ToolButtonsFromHBox collects tooltip buttons from an HBox container.
func ToolButtonsFromHBox(box *fyne.Container) []*ttwidget.Button {
	btns := make([]*ttwidget.Button, 0, len(box.Objects))
	for _, obj := range box.Objects {
		if btn := ToolButtonFrom(obj); btn != nil {
			btns = append(btns, btn)
		}
	}
	return btns
}
