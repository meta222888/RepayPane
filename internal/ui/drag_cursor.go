package ui

import (
	"github.com/relaypane/relaypane/internal/assets"

	"fyne.io/fyne/v2/driver/desktop"
)

func fileDragCursor() desktop.Cursor {
	return assets.FileDragCursor()
}

func applyFileDragCursor() {
	applyNativeFileDragCursor()
}

func clearFileDragCursor() {
	clearNativeFileDragCursor()
}
