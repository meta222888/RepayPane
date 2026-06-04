//go:build !windows

package ui

import "fyne.io/fyne/v2"

func winBeginDrag(*dragRegion) bool       { return false }
func winContinueDrag(*dragRegion)         {}
func winEndDrag()                          {}
func winLeftButtonDown() bool               { return false }
func winMinimizeWindows(fyne.Window)      {}
func winMaximizeWindows(fyne.Window)      {}
func winRestoreWindows(fyne.Window)       {}
