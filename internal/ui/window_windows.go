//go:build windows

package ui

import (
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

var user32 = syscall.NewLazyDLL("user32.dll")

const swMaximize = 3

func MaximizeWindow(w fyne.Window) {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		w.Resize(fyne.NewSize(1280, 800))
		w.CenterOnScreen()
		return
	}

	showWindow := user32.NewProc("ShowWindow")
	nw.RunNative(func(ctx any) {
		win, ok := ctx.(driver.WindowsWindowContext)
		if !ok || win.HWND == 0 {
			return
		}
		showWindow.Call(win.HWND, uintptr(swMaximize))
	})
}
