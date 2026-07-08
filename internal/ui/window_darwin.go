//go:build darwin

package ui

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit
#import <AppKit/AppKit.h>

void maximizeWindow(void *win) {
    NSWindow *window = (NSWindow*)win;
    if (window == nil) {
        return;
    }
    NSRect frame = [[window screen] visibleFrame];
    [window setFrame:frame display:YES animate:NO];
}
*/
import "C"
import (
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

func MaximizeWindow(w fyne.Window) {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		w.Resize(fyne.NewSize(1280, 800))
		w.CenterOnScreen()
		return
	}

	nw.RunNative(func(ctx any) {
		mac, ok := ctx.(driver.MacWindowContext)
		if !ok || mac.NSWindow == 0 {
			return
		}
		C.maximizeWindow(unsafe.Pointer(mac.NSWindow))
	})
}
