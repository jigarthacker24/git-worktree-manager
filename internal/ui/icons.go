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

func PinIcon() fyne.Resource {
	return theme.NewThemedResource(fyne.NewStaticResource("pin.svg", pinIcon))
}

func PinFilledIcon() fyne.Resource {
	return theme.NewThemedResource(fyne.NewStaticResource("pin-filled.svg", pinFilledIcon))
}
