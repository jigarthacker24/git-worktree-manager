package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// IconSelectOption is one choice in an IconSelect.
type IconSelectOption struct {
	Label string
	Icon  fyne.Resource
}

// IconSelect is a select control that shows an icon beside the selected label.
type IconSelect struct {
	widget.DisableableWidget

	options   []IconSelectOption
	selected  string
	OnChanged func(string)

	hovered bool
	popUp   *widget.PopUpMenu
}

// NewIconSelect creates a select with icon-labelled options.
func NewIconSelect(options []IconSelectOption, onChanged func(string)) *IconSelect {
	s := &IconSelect{
		options:   options,
		OnChanged: onChanged,
	}
	s.ExtendBaseWidget(s)
	return s
}

// SetSelected sets the current option by label.
func (s *IconSelect) SetSelected(label string) {
	for _, opt := range s.options {
		if opt.Label == label {
			s.selected = label
			s.Refresh()
			return
		}
	}
}

// Selected returns the current option label.
func (s *IconSelect) Selected() string {
	return s.selected
}

func (s *IconSelect) CreateRenderer() fyne.WidgetRenderer {
	th := s.Theme()
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.StrokeColor = theme.Color(theme.ColorNameInputBorder)
	bg.StrokeWidth = theme.InputBorderSize()
	bg.CornerRadius = th.Size(theme.SizeNameInputRadius)

	icon := widget.NewIcon(nil)
	label := widget.NewLabel(s.selected)
	label.Truncation = fyne.TextTruncateEllipsis
	arrow := widget.NewIcon(th.Icon(theme.IconNameArrowDropDown))

	r := &iconSelectRenderer{
		owner: s,
		bg:    bg,
		icon:  icon,
		label: label,
		arrow: arrow,
	}
	r.updateContent()
	return r
}

func (s *IconSelect) Tapped(*fyne.PointEvent) {
	if s.Disabled() {
		return
	}
	s.showPopUp()
}

func (s *IconSelect) showPopUp() {
	items := make([]*fyne.MenuItem, len(s.options))
	for i, opt := range s.options {
		label := opt.Label
		items[i] = &fyne.MenuItem{
			Label: label,
			Icon:  opt.Icon,
			Action: func() {
				s.setSelected(label)
				s.popUp = nil
			},
		}
	}

	c := fyne.CurrentApp().Driver().CanvasForObject(s)
	pop := widget.NewPopUpMenu(fyne.NewMenu("", items...), c)
	pop.ShowAtPosition(s.popUpPos())
	pop.Resize(fyne.NewSize(s.Size().Width, pop.MinSize().Height))
	pop.OnDismiss = func() {
		pop.Hide()
		if s.popUp == pop {
			s.popUp = nil
		}
	}
	s.popUp = pop
}

func (s *IconSelect) popUpPos() fyne.Position {
	buttonPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(s)
	return buttonPos.Add(fyne.NewPos(0, s.Size().Height-s.Theme().Size(theme.SizeNameInputBorder)))
}

func (s *IconSelect) setSelected(label string) {
	if s.selected == label {
		return
	}
	s.selected = label
	s.Refresh()
	if s.OnChanged != nil {
		s.OnChanged(label)
	}
}

func (s *IconSelect) MouseIn(*desktop.MouseEvent) {
	if s.Disabled() {
		return
	}
	s.hovered = true
	s.Refresh()
}

func (s *IconSelect) MouseOut() {
	s.hovered = false
	s.Refresh()
}

func (s *IconSelect) MouseMoved(*desktop.MouseEvent) {}

// DoubleClickEditorOptions returns the default-editor choices with icons.
func DoubleClickEditorOptions() []IconSelectOption {
	return []IconSelectOption{
		{Label: "None", Icon: theme.CancelIcon()},
		{Label: "VS Code", Icon: VSCodeIcon()},
		{Label: "Cursor", Icon: CursorIcon()},
		{Label: "Claude Code", Icon: ClaudeIcon()},
	}
}

type iconSelectRenderer struct {
	owner *IconSelect
	bg    *canvas.Rectangle
	icon  *widget.Icon
	label *widget.Label
	arrow *widget.Icon
}

func (r *iconSelectRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.icon, r.label, r.arrow}
}

func (r *iconSelectRenderer) Destroy() {}

func (r *iconSelectRenderer) Layout(size fyne.Size) {
	th := r.owner.Theme()
	pad := th.Size(theme.SizeNamePadding)
	innerPad := th.Size(theme.SizeNameInnerPadding)
	iconSize := th.Size(theme.SizeNameInlineIcon)

	r.bg.Resize(size)

	iconPos := fyne.NewPos(pad+innerPad, (size.Height-iconSize)/2)
	r.icon.Resize(fyne.NewSquareSize(iconSize))
	r.icon.Move(iconPos)

	arrowPos := fyne.NewPos(size.Width-iconSize-innerPad, (size.Height-iconSize)/2)
	r.arrow.Resize(fyne.NewSquareSize(iconSize))
	r.arrow.Move(arrowPos)

	labelX := iconPos.X + iconSize + pad
	labelW := arrowPos.X - labelX - pad
	labelH := r.label.MinSize().Height
	r.label.Resize(fyne.NewSize(labelW, labelH))
	r.label.Move(fyne.NewPos(labelX, (size.Height-labelH)/2))
}

func (r *iconSelectRenderer) MinSize() fyne.Size {
	th := r.owner.Theme()
	pad := th.Size(theme.SizeNamePadding)
	innerPad := th.Size(theme.SizeNameInnerPadding)
	iconSize := th.Size(theme.SizeNameInlineIcon)

	width := pad*2 + innerPad*2 + iconSize*2 + pad + r.label.MinSize().Width
	height := r.label.MinSize().Height + pad*2
	minH := th.Size(theme.SizeNameInputBorder)
	if height < minH {
		height = minH
	}
	return fyne.NewSize(width, height)
}

func (r *iconSelectRenderer) Refresh() {
	th := r.owner.Theme()

	if r.owner.Disabled() {
		r.bg.FillColor = theme.Color(theme.ColorNameInputBackground)
	} else if r.owner.hovered {
		r.bg.FillColor = theme.Color(theme.ColorNameHover)
	} else {
		r.bg.FillColor = theme.Color(theme.ColorNameInputBackground)
	}
	r.bg.StrokeColor = theme.Color(theme.ColorNameInputBorder)
	r.bg.CornerRadius = th.Size(theme.SizeNameInputRadius)

	r.updateContent()
	r.arrow.SetResource(th.Icon(theme.IconNameArrowDropDown))
	r.label.Importance = widget.MediumImportance

	r.bg.Refresh()
	r.icon.Refresh()
	r.label.Refresh()
	r.arrow.Refresh()
}

func (r *iconSelectRenderer) updateContent() {
	var icon fyne.Resource
	for _, opt := range r.owner.options {
		if opt.Label == r.owner.selected {
			icon = opt.Icon
			break
		}
	}
	r.icon.SetResource(icon)
	r.label.SetText(r.owner.selected)
}

var (
	_ fyne.Tappable       = (*IconSelect)(nil)
	_ desktop.Hoverable   = (*IconSelect)(nil)
	_ fyne.Disableable    = (*IconSelect)(nil)
)
