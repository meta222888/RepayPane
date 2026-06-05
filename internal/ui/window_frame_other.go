//go:build !windows

package ui

import "fyne.io/fyne/v2"

func fyneMoveWindowBy(fyne.Window, int, int) bool { return false }

func winBeginDrag(fyne.Window) bool { return false }
func winContinueDrag(*dragRegion)    {}
func winEndDrag()                     {}
func winLeftButtonDown() bool         { return false }
func winMinimizeWindows(fyne.Window) {}
func winMaximizeWindows(fyne.Window) {}
func winRestoreWindows(fyne.Window)  {}
func winInstallResizeHook(fyne.Window) bool { return false }
func winToggleMaximize(fyne.Window)  {}
func winIsMaximized(fyne.Window) bool { return false }

func wrapWindowResizePlatform(w fyne.Window, content fyne.CanvasObject) fyne.CanvasObject {
	return content
}
