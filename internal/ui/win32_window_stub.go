//go:build !windows

package ui

import "fyne.io/fyne/v2"

func setWindowOwner(child, owner fyne.Window) {}

func raiseWindow(w fyne.Window) {}
