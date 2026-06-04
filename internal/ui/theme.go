package ui

import (
	"image/color"
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

// RelayPane dark palette — aligned with Lovable reference (styles.css oklch).
var (
	colorBG           = color.NRGBA{R: 38, G: 41, B: 50, A: 255}   // background
	colorPanelHeader  = color.NRGBA{R: 46, G: 49, B: 58, A: 255}   // panel-header
	colorPanel        = color.NRGBA{R: 42, G: 45, B: 54, A: 255}
	colorBorder       = color.NRGBA{R: 58, G: 62, B: 72, A: 255}
	colorInput        = color.NRGBA{R: 48, G: 51, B: 60, A: 255}
	colorAccent       = color.NRGBA{R: 0, G: 180, B: 204, A: 255}   // primary cyan button (#00B4CC)
	colorAccentText   = color.NRGBA{R: 10, G: 10, B: 10, A: 255}    // black text on accent
	colorForeground   = color.NRGBA{R: 232, G: 234, B: 240, A: 255}
	colorMuted        = color.NRGBA{R: 148, G: 154, B: 168, A: 255}
	colorConnected    = color.NRGBA{R: 74, G: 200, B: 130, A: 255} // success
	colorDisconnected = color.NRGBA{R: 120, G: 126, B: 138, A: 255}
	colorTabActive    = color.NRGBA{R: 26, G: 72, B: 82, A: 255}   // same family as selection
	colorTabInactive  = color.NRGBA{R: 38, G: 41, B: 50, A: 255}
	colorHover        = color.NRGBA{R: 55, G: 60, B: 72, A: 255}
	colorRowSelected  = color.NRGBA{R: 26, G: 72, B: 82, A: 255}   // dark teal selection (#1A4852)
	colorRowHover     = color.NRGBA{R: 50, G: 54, B: 64, A: 255}
	colorWarning      = color.NRGBA{R: 220, G: 180, B: 90, A: 255}
)

func withBorderBottom(obj fyne.CanvasObject) fyne.CanvasObject {
	line := canvas.NewRectangle(colorBorder)
	line.SetMinSize(fyne.NewSize(0, 1))
	return container.NewBorder(nil, line, nil, nil, withBackground(obj, colorPanelHeader))
}

func withBackground(obj fyne.CanvasObject, bg color.Color) fyne.CanvasObject {
	rect := canvas.NewRectangle(bg)
	return container.NewStack(rect, container.NewPadded(obj))
}

func withPanelHeader(obj fyne.CanvasObject) fyne.CanvasObject {
	return withBorderBottom(obj)
}

func withStatusBar(obj fyne.CanvasObject) fyne.CanvasObject {
	top := canvas.NewRectangle(colorBorder)
	top.SetMinSize(fyne.NewSize(0, 1))
	return container.NewBorder(top, nil, nil, nil, withBackground(obj, colorPanelHeader))
}

func withPanelLabel(obj fyne.CanvasObject) fyne.CanvasObject {
	return withBorderBottom(withBackground(obj, colorBG))
}

func splitBorder(obj fyne.CanvasObject) fyne.CanvasObject {
	line := canvas.NewRectangle(colorBorder)
	line.SetMinSize(fyne.NewSize(1, 0))
	return container.NewBorder(nil, nil, line, nil, obj)
}

func fixedWidth(obj fyne.CanvasObject, width float32) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(width, 0))
	return container.NewStack(spacer, obj)
}

type relayPaneTheme struct {
	font fyne.Resource
}

func newRelayPaneTheme() *relayPaneTheme {
	return &relayPaneTheme{font: loadSystemFont()}
}

func loadSystemFont() fyne.Resource {
	if runtime.GOOS == "windows" {
		windir := os.Getenv("WINDIR")
		if windir == "" {
			windir = `C:\Windows`
		}
		for _, name := range []string{"msyh.ttc", "msyhbd.ttc", "segoeui.ttf"} {
			p := filepath.Join(windir, "Fonts", name)
			if res, err := fyne.LoadResourceFromPath(p); err == nil {
				return res
			}
		}
	}
	return theme.DefaultTheme().Font(fyne.TextStyle{})
}

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
	case theme.ColorNameForegroundOnPrimary:
		return colorAccentText
	case theme.ColorNameHover:
		return colorHover
	case theme.ColorNameInputBackground:
		return colorInput
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
		return colorRowSelected
	case theme.ColorNameSeparator:
		return colorBorder
	}
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (t *relayPaneTheme) Font(style fyne.TextStyle) fyne.Resource {
	if t.font != nil {
		return t.font
	}
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
		return 11
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 4
	case theme.SizeNameInlineIcon:
		return 14
	}
	return theme.DefaultTheme().Size(name)
}

func ApplyTheme(a fyne.App) {
	a.Settings().SetTheme(newRelayPaneTheme())
}
