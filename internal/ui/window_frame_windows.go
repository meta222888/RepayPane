//go:build windows

package ui

import (
	"syscall"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	procShowWindow   = user32.NewProc("ShowWindow")
	procGetWindowRect = user32.NewProc("GetWindowRect")
	procSetWindowPos = user32.NewProc("SetWindowPos")
)

const (
	swMinimize  = 6
	swMaximize  = 3
	swRestore   = 9
	swpNoSize   = 0x0001
	swpNoZOrder = 0x0004
)

type winRect struct {
	Left, Top, Right, Bottom int32
}

func winMoveWindows(w fyne.Window, scale float32, delta fyne.Delta) {
	dx := int(delta.DX * scale)
	dy := int(delta.DY * scale)
	if dx == 0 && dy == 0 {
		return
	}
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return
	}
	nw.RunNative(func(ctx any) {
		c, ok := ctx.(driver.WindowsWindowContext)
		if !ok || c.HWND == 0 {
			return
		}
		var r winRect
		procGetWindowRect.Call(c.HWND, uintptr(unsafe.Pointer(&r)))
		procSetWindowPos.Call(
			c.HWND, 0,
			uintptr(int64(r.Left)+int64(dx)),
			uintptr(int64(r.Top)+int64(dy)),
			0, 0,
			swpNoSize|swpNoZOrder,
		)
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
