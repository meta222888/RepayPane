package ui

import (
	"image/color"
	"os"
	"path/filepath"
	"runtime"

	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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
	// Semi-transparent overlays for button hover/focus — must not be opaque or accent turns gray.
	colorHover          = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x22}
	colorFocus          = color.NRGBA{R: 0x00, G: 0xc8, B: 0xb4, A: 0x40}
	colorPressed        = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x28}
	colorRowSelected    = color.NRGBA{R: 0x00, G: 0x6e, B: 0x63, A: 0xcc}
	colorRowAlt         = color.NRGBA{R: 0x22, G: 0x26, B: 0x2f, A: 255}
	colorRowHover       = color.NRGBA{R: 0x2a, G: 0x30, B: 0x3d, A: 255}
	colorTextHighlight  = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 255}
	colorStatusBar      = color.NRGBA{R: 0x13, G: 0x16, B: 0x1b, A: 255}
	colorWarning        = color.NRGBA{R: 0xff, G: 0x8c, B: 0x42, A: 255}
)

const (
	textDescenderPad     float32 = 4
	AppTextSize          float32 = 12
	AppTitleTextSize     float32 = 14
	PaneRowHeight        float32 = 28
	paneRowNameSize      float32 = AppTextSize
	paneRowMetaSize      float32 = AppTextSize
	paneRowPadH          float32 = 4
	paneRowPadV          float32 = 2
	topBarHeight         float32 = 28
	statusBarHeight      float32 = 22
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
	bg := canvas.NewRectangle(colorPanelHeader)
	bg.SetMinSize(fyne.NewSize(0, topBarHeight))
	line := canvas.NewRectangle(colorBorder)
	line.SetMinSize(fyne.NewSize(0, 1))
	padded := container.New(layout.NewCustomPaddedLayout(6, 2, 8, 2), obj)
	return container.NewVBox(
		container.NewStack(bg, padded),
		line,
	)
}

func withStatusBar(obj fyne.CanvasObject) fyne.CanvasObject {
	top := canvas.NewRectangle(colorBorder)
	top.SetMinSize(fyne.NewSize(0, 1))
	bg := canvas.NewRectangle(colorStatusBar)
	bg.SetMinSize(fyne.NewSize(0, statusBarHeight))
	padded := container.New(layout.NewCustomPaddedLayout(6, 2, 8, 2), obj)
	inner := container.NewStack(bg, padded)
	return container.NewBorder(top, nil, nil, nil, inner)
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

func labelCText(text string, c color.Color, size float32) fyne.CanvasObject {
	if size < 1 {
		size = AppTextSize
	}
	t := canvas.NewText(text, c)
	t.TextSize = size
	return wrapCanvasText(t)
}

func labelCBandText(text string, c color.Color, size float32) fyne.CanvasObject {
	if size < 1 {
		size = AppTextSize
	}
	t := canvas.NewText(text, c)
	t.TextSize = size
	sz, _ := fyne.CurrentApp().Driver().RenderedTextSize(t.Text, t.TextSize, t.TextStyle, t.FontSource)
	if sz.Height < size {
		sz.Height = size
	}
	if sz.Width < 1 {
		sz.Width = 1
	}
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(sz)
	return container.NewStack(spacer, t)
}

func titleLabel(text string) fyne.CanvasObject {
	lbl := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	return container.NewThemeOverride(lbl, newTitleTextTheme())
}

func titleLabelAlign(text string, align fyne.TextAlign) fyne.CanvasObject {
	lbl := widget.NewLabelWithStyle(text, align, fyne.TextStyle{Bold: true})
	return container.NewThemeOverride(lbl, newTitleTextTheme())
}

func wrapTitleLabel(lbl *widget.Label) fyne.CanvasObject {
	return container.NewThemeOverride(lbl, newTitleTextTheme())
}

// wrapCanvasText reserves vertical space so Latin descenders (g, j, p, y) are not clipped.
func wrapCanvasText(t *canvas.Text) fyne.CanvasObject {
	sz, _ := fyne.CurrentApp().Driver().RenderedTextSize(t.Text, t.TextSize, t.TextStyle, t.FontSource)
	if sz.Height < t.TextSize+textDescenderPad {
		sz.Height = t.TextSize + textDescenderPad
	} else {
		sz.Height += 2
	}
	if sz.Width < 1 {
		sz.Width = 1
	}
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(sz)
	return container.NewStack(spacer, t)
}

func bandPadding(content fyne.CanvasObject) fyne.CanvasObject {
	return container.New(layout.NewCustomPaddedLayout(paneRowPadH, paneRowPadV, paneRowPadH, paneRowPadV), content)
}

func panelBand(content fyne.CanvasObject, height float32) fyne.CanvasObject {
	bg := canvas.NewRectangle(colorPanelHeader)
	bg.SetMinSize(fyne.NewSize(0, height))
	line := canvas.NewRectangle(colorBorder)
	line.SetMinSize(fyne.NewSize(0, 1))
	return container.NewVBox(
		container.NewStack(bg, bandPadding(content)),
		line,
	)
}

// paddedWidgetLabel wraps a label with room for descenders (g, j, p, y).
func paddedWidgetLabel(lbl *widget.Label) fyne.CanvasObject {
	return container.New(layout.NewCustomPaddedLayout(0, 0, 0, 3), lbl)
}

func emptyPaneSlot() fyne.CanvasObject {
	s := canvas.NewRectangle(color.Transparent)
	s.SetMinSize(fyne.NewSize(0, 1))
	return s
}

type relayPaneTheme struct {
	regular fyne.Resource
	bold    fyne.Resource
	mono    fyne.Resource
}

type uiFontSet struct {
	regular, bold, mono string
}

func newRelayPaneTheme() *relayPaneTheme {
	regular, bold, mono := loadUIFonts(i18n.Current())
	return &relayPaneTheme{regular: regular, bold: bold, mono: mono}
}

func loadUIFonts(lang i18n.Lang) (regular, bold, mono fyne.Resource) {
	def := theme.DefaultTheme()
	if runtime.GOOS != "windows" {
		return def.Font(fyne.TextStyle{}), def.Font(fyne.TextStyle{Bold: true}), def.Font(fyne.TextStyle{Monospace: true})
	}
	windir := os.Getenv("WINDIR")
	if windir == "" {
		windir = `C:\Windows`
	}
	fontDir := filepath.Join(windir, "Fonts")

	// Prefer Segoe UI first — correct Latin descenders; includes CJK on modern Windows.
	for _, set := range englishFontCandidates() {
		if reg, bld, mon, ok := tryLoadFontSet(fontDir, set); ok {
			return reg, bld, mon
		}
	}
	if lang == i18n.ZH {
		for _, set := range chineseFontCandidates() {
			if reg, bld, mon, ok := tryLoadFontSet(fontDir, set); ok {
				return reg, bld, mon
			}
		}
	}
	return def.Font(fyne.TextStyle{}), def.Font(fyne.TextStyle{Bold: true}), def.Font(fyne.TextStyle{Monospace: true})
}

func tryLoadFontSet(fontDir string, set uiFontSet) (regular, bold, mono fyne.Resource, ok bool) {
	reg, err := fyne.LoadResourceFromPath(filepath.Join(fontDir, set.regular))
	if err != nil {
		return nil, nil, nil, false
	}
	bld := reg
	if set.bold != "" {
		if res, err := fyne.LoadResourceFromPath(filepath.Join(fontDir, set.bold)); err == nil {
			bld = res
		}
	}
	mon := bld
	if set.mono != "" {
		if res, err := fyne.LoadResourceFromPath(filepath.Join(fontDir, set.mono)); err == nil {
			mon = res
		}
	}
	return reg, bld, mon, true
}

func englishFontCandidates() []uiFontSet {
	return []uiFontSet{
		{regular: "segoeui.ttf", bold: "segoeuib.ttf", mono: "CascadiaMono.ttf"},
		{regular: "segoeuisl.ttf", bold: "segoeuib.ttf", mono: "CascadiaMono.ttf"},
		{regular: "arial.ttf", bold: "arialbd.ttf", mono: "consola.ttf"},
	}
}

func chineseFontCandidates() []uiFontSet {
	return []uiFontSet{
		{regular: "HarmonyOS_Sans_SC_Regular.ttf", bold: "HarmonyOS_Sans_SC_Bold.ttf", mono: "CascadiaMono.ttf"},
		{regular: "Deng.ttf", bold: "Dengb.ttf", mono: "CascadiaMono.ttf"},
		// simhei has poor Latin glyphs; keep only as last-resort for CJK coverage.
		{regular: "simhei.ttf", bold: "simhei.ttf", mono: "CascadiaMono.ttf"},
	}
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
	case theme.ColorNameFocus:
		return colorFocus
	case theme.ColorNamePressed:
		return colorPressed
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
	if style.Monospace && t.mono != nil {
		return t.mono
	}
	if style.Bold && t.bold != nil {
		return t.bold
	}
	if t.regular != nil {
		return t.regular
	}
	return theme.DefaultTheme().Font(style)
}

func (relayPaneTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (relayPaneTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return AppTextSize
	case theme.SizeNameCaptionText:
		return AppTextSize
	case theme.SizeNameLineSpacing:
		return 6
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 6
	case theme.SizeNameInlineIcon:
		return 16
	}
	return theme.DefaultTheme().Size(name)
}

// listCompactTheme removes inter-row padding from widget.List for dense file rows.
type listCompactTheme struct {
	base fyne.Theme
}

func newListCompactTheme() fyne.Theme {
	return &listCompactTheme{base: newRelayPaneTheme()}
}

func (t *listCompactTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}

func (t *listCompactTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *listCompactTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *listCompactTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNamePadding {
		return 0
	}
	return t.base.Size(name)
}

type compactToolbarTheme struct {
	base fyne.Theme
}

func newCompactToolbarTheme() fyne.Theme {
	return &compactToolbarTheme{base: fyne.CurrentApp().Settings().Theme()}
}

func (t *compactToolbarTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}

func (t *compactToolbarTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *compactToolbarTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *compactToolbarTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return AppTextSize
	case theme.SizeNamePadding:
		return 2
	case theme.SizeNameInnerPadding:
		return 2
	case theme.SizeNameInlineIcon:
		return 14
	}
	return t.base.Size(name)
}

const topMenuButtonMinWidth float32 = 88

// menuItemInnerPadding is applied via Fyne SizeNameInnerPadding inside each menu row
// (row height = text + inner padding). Do not simulate row spacing with outer margins.
const menuItemInnerPadding float32 = 6

type menuPopupTheme struct {
	base fyne.Theme
}

func newMenuPopupTheme() fyne.Theme {
	return &menuPopupTheme{base: fyne.CurrentApp().Settings().Theme()}
}

func (t *menuPopupTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}

func (t *menuPopupTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *menuPopupTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *menuPopupTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return AppTextSize
	case theme.SizeNameInnerPadding:
		return menuItemInnerPadding
	case theme.SizeNamePadding:
		return 0
	}
	return t.base.Size(name)
}

func wrapTopMenuButton(btn *widget.Button) fyne.CanvasObject {
	minW := canvas.NewRectangle(color.Transparent)
	minW.SetMinSize(fyne.NewSize(topMenuButtonMinWidth, 0))
	themed := container.NewThemeOverride(btn, newCompactToolbarTheme())
	return container.NewStack(minW, themed)
}

type paneChromeToolbarTheme struct {
	base fyne.Theme
}

func newPaneChromeToolbarTheme() fyne.Theme {
	return &paneChromeToolbarTheme{base: fyne.CurrentApp().Settings().Theme()}
}

func (t *paneChromeToolbarTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}

func (t *paneChromeToolbarTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *paneChromeToolbarTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *paneChromeToolbarTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return AppTextSize
	case theme.SizeNamePadding:
		return 1
	case theme.SizeNameInnerPadding:
		return 1
	case theme.SizeNameInlineIcon:
		return 14
	}
	return t.base.Size(name)
}

type paneChromeEntryTheme struct {
	base     fyne.Theme
	textSize float32
}

func newPaneChromeEntryTheme(textSize float32) fyne.Theme {
	if textSize < 1 {
		textSize = paneRowNameSize
	}
	return &paneChromeEntryTheme{
		base:     fyne.CurrentApp().Settings().Theme(),
		textSize: textSize,
	}
}

func newPanePathEntryTheme(textSize float32) fyne.Theme {
	if textSize < 1 {
		textSize = paneRowNameSize
	}
	return &panePathEntryTheme{
		base:     newRelayPaneTheme(),
		textSize: textSize,
	}
}

func (t *paneChromeEntryTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}

func (t *paneChromeEntryTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *paneChromeEntryTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *paneChromeEntryTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return t.textSize
	case theme.SizeNamePadding:
		return 1
	case theme.SizeNameInnerPadding:
		return 1
	}
	return t.base.Size(name)
}

type panePathEntryTheme struct {
	base     fyne.Theme
	textSize float32
}

func (t *panePathEntryTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameInputBackground:
		return colorInput
	case theme.ColorNameInputBorder:
		return colorBorder
	}
	return t.base.Color(name, variant)
}

func (t *panePathEntryTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *panePathEntryTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *panePathEntryTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return t.textSize
	case theme.SizeNamePadding:
		return 12
	case theme.SizeNameInnerPadding:
		return 3
	}
	return t.base.Size(name)
}

type compactEntryTheme struct {
	base     fyne.Theme
	textSize float32
}

func newCompactEntryTheme(textSize float32) fyne.Theme {
	if textSize < 1 {
		textSize = AppTextSize
	}
	return &compactEntryTheme{
		base:     fyne.CurrentApp().Settings().Theme(),
		textSize: textSize,
	}
}

func (t *compactEntryTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}

func (t *compactEntryTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *compactEntryTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *compactEntryTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return t.textSize
	case theme.SizeNamePadding:
		return 2
	case theme.SizeNameInnerPadding:
		return 2
	}
	return t.base.Size(name)
}

type titleTextTheme struct {
	base fyne.Theme
}

func newTitleTextTheme() fyne.Theme {
	return &titleTextTheme{base: fyne.CurrentApp().Settings().Theme()}
}

func (t *titleTextTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}

func (t *titleTextTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *titleTextTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *titleTextTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return AppTitleTextSize
	}
	return t.base.Size(name)
}

func ApplyTheme(a fyne.App) {
	a.Settings().SetTheme(newRelayPaneTheme())
}
