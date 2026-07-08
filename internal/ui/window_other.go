//go:build !darwin && !windows

package ui

import "fyne.io/fyne/v2"

func MaximizeWindow(w fyne.Window) {
	w.Resize(fyne.NewSize(1280, 800))
	w.CenterOnScreen()
}
