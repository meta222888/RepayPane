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

// RelayPane dark palette — aligned with relaypane-source Fyne shell.
var (
	colorBG             = color.NRGBA{R: 0x1a, G: 0x1d, B: 0x23, A: 255}
	colorPanel          = color.NRGBA{R: 0x1e, G: 0x22, B: 0x2a, A: 255}
	colorPanelHeader    = color.NRGBA{R: 0x16, G: 0x19, B: 0x1f, A: 255}
	colorBorder         = color.NRGBA{R: 0x2f, G: 0x33, B: 0x3d, A: 255}
	colorInput          = color.NRGBA{R: 0x20, G: 0x24, B: 0x2c, A: 255}
	colorAccent         = color.NRGBA{R: 0x00, G: 0xc8, B: 0xb4, A: 255}
	colorAccentText     = color.NRGBA{R: 0x0a, G: 0x0a, B: 0x0a, A: 255}
	colorForeground     = color.NRGBA{R: 0xd0, G: 0xd4, B: 0xdc, A: 255}
	colorMuted          = color.NRGBA{R: 0x9a, G: 0xa0, B: 0xb0, A: 255}
	colorConnected      = color.NRGBA{R: 0x3c, G: 0xd6, B: 0x68, A: 255}
	colorDisconnected   = color.NRGBA{R: 0x6b, G: 0x72, B: 0x80, A: 255}
	colorTabActive      = color.NRGBA{R: 0x1a, G: 0x1d, B: 0x23, A: 255}
	colorTabInactive    = color.NRGBA{R: 0x22, G: 0x26, B: 0x2f, A: 255}
	colorHover          = color.NRGBA{R: 0x28, G: 0x2c, B: 0x36, A: 255}
	colorRowSelected    = color.NRGBA{R: 0x00, G: 0x6e, B: 0x63, A: 0xcc}
	colorRowAlt         = color.NRGBA{R: 0x22, G: 0x26, B: 0x2f, A: 255}
	colorRowHover       = color.NRGBA{R: 0x2a, G: 0x30, B: 0x3d, A: 255}
	colorTextHighlight  = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 255}
	colorStatusBar      = color.NRGBA{R: 0x13, G: 0x16, B: 0x1b, A: 255}
	colorWarning        = color.NRGBA{R: 0xff, G: 0x8c, B: 0x42, A: 255}
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
	return container.NewBorder(top, nil, nil, nil, withBackground(obj, colorStatusBar))
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

func dotWidget(dot *canvas.Circle, size float32) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(size, size))
	return container.NewStack(spacer, container.NewCenter(dot))
}

func labelCText(text string, c color.Color, size float32) *canvas.Text {
	t := canvas.NewText(text, c)
	t.TextSize = size
	return t
}

func panelBand(content fyne.CanvasObject, height float32) fyne.CanvasObject {
	bg := canvas.NewRectangle(colorPanelHeader)
	bg.SetMinSize(fyne.NewSize(0, height))
	line := canvas.NewRectangle(colorBorder)
	line.SetMinSize(fyne.NewSize(0, 1))
	return container.NewVBox(
		container.NewStack(bg, container.NewPadded(content)),
		line,
	)
}

func emptyPaneSlot() fyne.CanvasObject {
	s := canvas.NewRectangle(color.Transparent)
	s.SetMinSize(fyne.NewSize(0, 1))
	return s
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
		// Fyne cannot use .ttc font collections — only single .ttf files.
		for _, name := range []string{"msyh.ttf", "simhei.ttf", "deng.ttf", "segoeui.ttf", "arial.ttf"} {
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
		return color.NRGBA{R: 0x00, G: 0xc8, B: 0xb4, A: 0x66}
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
