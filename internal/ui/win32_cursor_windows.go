//go:build windows

package ui

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"

	"github.com/relaypane/relaypane/internal/assets"
)

const (
	imageCursor     = 2
	lrLoadFromFile  = 0x00000010
	lrDefaultSize   = 0x00000040
	idcArrow        = 32512
)

var (
	loadImageW = user32.NewProc("LoadImageW")
	setCursor  = user32.NewProc("SetCursor")
	loadCursorW = user32.NewProc("LoadCursorW")

	fileDragCursorHandle uintptr
	fileDragCursorOnce   sync.Once
)

func initNativeFileDragCursor() {
	fileDragCursorOnce.Do(func() {
		path := filepath.Join(os.TempDir(), "relaypane-file-drag.cur")
		if err := os.WriteFile(path, assets.CopyCURBytes(), 0o644); err != nil {
			return
		}
		pathW, err := syscall.UTF16PtrFromString(path)
		if err != nil {
			return
		}
		h, _, _ := loadImageW.Call(
			0,
			uintptr(unsafe.Pointer(pathW)),
			imageCursor,
			0,
			0,
			lrLoadFromFile|lrDefaultSize,
		)
		if h != 0 {
			fileDragCursorHandle = h
		}
	})
}

func applyNativeFileDragCursor() {
	initNativeFileDragCursor()
	if fileDragCursorHandle != 0 {
		setCursor.Call(fileDragCursorHandle)
	}
}

func clearNativeFileDragCursor() {
	h, _, _ := loadCursorW.Call(0, idcArrow)
	if h != 0 {
		setCursor.Call(h)
	}
}
