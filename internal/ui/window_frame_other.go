//go:build !windows

package ui

import (
	"time"

	"fyne.io/fyne/v2"
)

func doubleClickInterval() time.Duration { return 500 * time.Millisecond }

func winStartCaptionDrag(fyne.Window) bool { return false }

func winMinimizeWindows(fyne.Window) {}
func winMaximizeWindows(fyne.Window) {}
func winRestoreWindows(fyne.Window)  {}
func winInstallResizeHook(fyne.Window) bool { return false }
func winToggleMaximize(fyne.Window)  {}
func winIsMaximized(fyne.Window) bool { return false }

func wrapWindowResizePlatform(w fyne.Window, content fyne.CanvasObject) fyne.CanvasObject {
	return content
}
