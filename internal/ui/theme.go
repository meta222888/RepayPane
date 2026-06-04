package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// RelayPane dark palette aligned with file-compass reference UI.
var (
	colorBG          = color.NRGBA{R: 28, G: 31, B: 38, A: 255}
	colorPanelHeader = color.NRGBA{R: 34, G: 37, B: 45, A: 255}
	colorPanel       = color.NRGBA{R: 30, G: 33, B: 40, A: 255}
	colorBorder      = color.NRGBA{R: 50, G: 54, B: 62, A: 255}
	colorAccent      = color.NRGBA{R: 59, G: 130, B: 246, A: 255}
	colorForeground  = color.NRGBA{R: 232, G: 234, B: 237, A: 255}
	colorMuted       = color.NRGBA{R: 154, G: 160, B: 166, A: 255}
	colorConnected   = color.NRGBA{R: 34, G: 197, B: 94, A: 255}
	colorDisconnected = color.NRGBA{R: 107, G: 114, B: 128, A: 255}
	colorTabActive   = color.NRGBA{R: 42, G: 46, B: 56, A: 255}
	colorTabInactive = color.NRGBA{R: 28, G: 31, B: 38, A: 255}
	colorHover       = color.NRGBA{R: 55, G: 60, B: 72, A: 255}
)

type relayPaneTheme struct{}

func (relayPaneTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colorBG
	case theme.ColorNameButton:
		return colorPanelHeader
	case theme.ColorNameDisabledButton:
		return colorPanel
	case theme.ColorNameForeground:
		return colorForeground
	case theme.ColorNameDisabled:
		return colorMuted
	case theme.ColorNamePlaceHolder:
		return colorMuted
	case theme.ColorNamePrimary:
		return colorAccent
	case theme.ColorNameHover:
		return colorHover
	case theme.ColorNameInputBackground:
		return colorPanel
	case theme.ColorNameInputBorder:
		return colorBorder
	case theme.ColorNameScrollBar:
		return colorBorder
	case theme.ColorNameShadow:
		return color.NRGBA{A: 80}
	case theme.ColorNameHeaderBackground:
		return colorPanelHeader
	case theme.ColorNameMenuBackground:
		return colorPanelHeader
	case theme.ColorNameOverlayBackground:
		return colorPanelHeader
	case theme.ColorNameSelection:
		return color.NRGBA{R: 59, G: 130, B: 246, A: 90}
	case theme.ColorNameSeparator:
		return colorBorder
	}
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (relayPaneTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (relayPaneTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (relayPaneTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 13
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 4
	case theme.SizeNameInlineIcon:
		return 16
	}
	return theme.DefaultTheme().Size(name)
}

func ApplyTheme(a fyne.App) {
	a.Settings().SetTheme(&relayPaneTheme{})
}
