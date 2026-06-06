package walkui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// Layout sizes from public/design-tokens.json @ 96 DPI.
const (
	winDefaultW = 1280
	winDefaultH = 760
	winMinW     = 960
	winMinH     = 600

	tabBarHeight   = 24
	toolBarHeight  = 28
	navRowHeight   = 22
	paneTitleH     = 18
	statusBarH     = 28
	navIconSize    = 22
	markIconSize   = 14
	driveComboW    = 56
	placesComboW   = 100
	remoteSpacerW  = 184
	statusDotTab   = 6
	statusDotBar   = 8
	progressBarW   = 180
	progressBarH   = 10
	tabMaxWidth    = 160
)

var (
	colorWindowChrome = walk.RGB(240, 240, 240)
	colorTabActive    = walk.RGB(220, 235, 252)
	colorTabInactive  = walk.RGB(240, 240, 240)
	colorSelection    = walk.RGB(203, 232, 246)
	colorTextMuted    = walk.RGB(85, 85, 85)
	colorRemoteEmpty  = walk.RGB(136, 136, 136)
)

func uiFont() Font {
	return Font{Family: "Segoe UI", PointSize: 9}
}

func uiFontBold() Font {
	return Font{Family: "Segoe UI", PointSize: 9, Bold: true}
}

func monoFont() Font {
	return Font{Family: "Consolas", PointSize: 9}
}

func fixedHeight(h int) Size {
	return Size{Width: 0, Height: h}
}

func paneMargins() Margins {
	return Margins{Left: 6, Top: 6, Right: 6, Bottom: 0}
}

func dlgBody(children ...Widget) Widget {
	return Composite{
		Layout:   VBox{Margins: Margins{Left: 12, Top: 12, Right: 12, Bottom: 12}},
		Children: children,
	}
}

func dlgFooter(buttons ...Widget) Widget {
	row := append([]Widget{HSpacer{}}, buttons...)
	return Composite{
		Layout:   HBox{Margins: Margins{Left: 12, Top: 8, Right: 12, Bottom: 8}},
		Children: row,
	}
}

func navIconButton(bmp *walk.Bitmap, tooltip string, fn func()) Widget {
	return ToolButton{
		Image:       bmp,
		ToolTipText: tooltip,
		OnClicked:   fn,
		MinSize:     Size{navIconSize, navIconSize},
		MaxSize:     Size{navIconSize, navIconSize},
	}
}

func markIconView(bmp *walk.Bitmap) Widget {
	return ImageView{
		Image:   bmp,
		Mode:    ImageViewModeShrink,
		MinSize: Size{markIconSize, markIconSize},
		MaxSize: Size{markIconSize, markIconSize},
	}
}

func remoteNavSpacer() Widget {
	return Composite{
		MinSize: Size{remoteSpacerW, 0},
		MaxSize: Size{remoteSpacerW, 0},
		Layout:  HBox{MarginsZero: true},
	}
}

func tabStateColor(state tabState) walk.Color {
	switch state {
	case tabConnected:
		return colorConnected
	case tabConnecting:
		return colorConnecting
	default:
		return colorDisconnected
	}
}

func tableColumns() []TableViewColumn {
	return []TableViewColumn{
		{Title: "Name", Width: 280},
		{Title: "Size", Width: 80},
		{Title: "Modified", Width: 140},
	}
}
