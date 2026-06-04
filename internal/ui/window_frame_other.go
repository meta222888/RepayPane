//go:build !windows

package ui

import "fyne.io/fyne/v2"

func winMoveWindows(w fyne.Window, scale float32, delta fyne.Delta) {
	_ = w
	_ = scale
	_ = delta
}

func winMinimizeWindows(fyne.Window) {}
func winMaximizeWindows(fyne.Window) {}
func winRestoreWindows(fyne.Window)  {}
