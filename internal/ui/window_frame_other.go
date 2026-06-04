//go:build !windows

package ui

import "fyne.io/fyne/v2"

func winDragWindows(fyne.Window)     {}
func winMinimizeWindows(fyne.Window) {}
func winMaximizeWindows(fyne.Window) {}
func winRestoreWindows(fyne.Window)  {}
