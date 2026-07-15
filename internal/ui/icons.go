package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed icons/pin.svg
var pinIcon []byte

//go:embed icons/pin-filled.svg
var pinFilledIcon []byte

//go:embed icons/vscode.svg
var vscodeIcon []byte

//go:embed icons/cursor-dark.svg
var cursorIconDark []byte

//go:embed icons/cursor-light.svg
var cursorIconLight []byte

//go:embed icons/claude.svg
var claudeIcon []byte

func PinIcon() fyne.Resource {
	return theme.NewThemedResource(fyne.NewStaticResource("pin.svg", pinIcon))
}

func PinFilledIcon() fyne.Resource {
	return theme.NewThemedResource(fyne.NewStaticResource("pin-filled.svg", pinFilledIcon))
}

func VSCodeIcon() fyne.Resource {
	return fyne.NewStaticResource("vscode.svg", vscodeIcon)
}

func CursorIcon() fyne.Resource {
	data := cursorIconDark
	if app := fyne.CurrentApp(); app != nil && app.Settings().ThemeVariant() == theme.VariantDark {
		data = cursorIconLight
	}
	return fyne.NewStaticResource("cursor.svg", data)
}

func ClaudeIcon() fyne.Resource {
	return fyne.NewStaticResource("claude.svg", claudeIcon)
}
