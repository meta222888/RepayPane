//go:build windows

package ui

import (
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procReleaseCapture   = user32.NewProc("ReleaseCapture")
	procSendMessageW     = user32.NewProc("SendMessageW")
	procShowWindow       = user32.NewProc("ShowWindow")
)

const (
	wmNCLButtonDown = 0x00A1
	htCaption       = 2
	swMinimize      = 6
	swMaximize      = 3
	swRestore       = 9
)

func winDragWindows(w fyne.Window) {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return
	}
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if !ok || c.HWND == 0 {
			return
		}
		procReleaseCapture.Call()
		procSendMessageW.Call(c.HWND, wmNCLButtonDown, htCaption, 0)
	})
}

func winMinimizeWindows(w fyne.Window) {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return
	}
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if !ok || c.HWND == 0 {
			return
		}
		procShowWindow.Call(c.HWND, swMinimize)
	})
}

func winMaximizeWindows(w fyne.Window) {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return
	}
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if !ok || c.HWND == 0 {
			return
		}
		procShowWindow.Call(c.HWND, swMaximize)
	})
}

func winRestoreWindows(w fyne.Window) {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return
	}
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if !ok || c.HWND == 0 {
			return
		}
		procShowWindow.Call(c.HWND, swRestore)
	})
}
