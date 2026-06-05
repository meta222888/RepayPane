//go:build windows

package ui

import (
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	setWindowLongPtr = user32.NewProc("SetWindowLongPtrW")
	setWindowPos     = user32.NewProc("SetWindowPos")
)

const (
	gwlHWNDParent = ^uintptr(7) // -8
	swpNoActivate = 0x0010
	swpNoMove     = 0x0002
	swpNoSize     = 0x0001
	swpShowWindow = 0x0040
	hwndTop       = 0
)

func windowHWND(w fyne.Window) uintptr {
	if w == nil {
		return 0
	}
	var hwnd uintptr
	if nw, ok := w.(driver.NativeWindow); ok {
		nw.RunNative(func(ctx any) {
			if c, ok := ctx.(driver.WindowsWindowContext); ok {
				hwnd = c.HWND
			}
		})
	}
	return hwnd
}

func setWindowOwner(child, owner fyne.Window) {
	childHWND := windowHWND(child)
	ownerHWND := windowHWND(owner)
	if childHWND == 0 || ownerHWND == 0 || childHWND == ownerHWND {
		return
	}
	setWindowLongPtr.Call(childHWND, gwlHWNDParent, ownerHWND)
}

func raiseWindow(w fyne.Window) {
	hwnd := windowHWND(w)
	if hwnd == 0 {
		return
	}
	setWindowPos.Call(
		hwnd,
		hwndTop,
		0,
		0,
		0,
		0,
		uintptr(swpNoMove|swpNoSize|swpNoActivate|swpShowWindow),
	)
}
