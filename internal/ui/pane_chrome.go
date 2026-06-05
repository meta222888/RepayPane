package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	paneBandLineH    float32 = 1
	paneIconBtnSize  float32 = 20
	paneIconGlyphSize float32 = 14
)

func paneBandInnerHeight() float32 {
	return PaneRowHeight - paneBandLineH - paneRowPadV*2
}

// paneFixedHeight reports a fixed height so Entry/Button widgets cannot stretch chrome rows.
type paneFixedHeight struct {
	widget.BaseWidget
	child  fyne.CanvasObject
	height float32
}

func newPaneFixedHeight(height float32, child fyne.CanvasObject) *paneFixedHeight {
	f := &paneFixedHeight{child: child, height: height}
	f.ExtendBaseWidget(f)
	return f
}

func (f *paneFixedHeight) MinSize() fyne.Size {
	w := f.child.MinSize().Width
	if w < 1 {
		w = 1
	}
	return fyne.NewSize(w, f.height)
}

type paneFixedHeightRenderer struct {
	f     *paneFixedHeight
	child fyne.CanvasObject
}

func (r *paneFixedHeightRenderer) Layout(size fyne.Size) {
	size.Height = r.f.height
	r.f.Resize(size)
	r.child.Resize(size)
	r.child.Move(fyne.NewPos(0, 0))
}

func (r *paneFixedHeightRenderer) MinSize() fyne.Size {
	return r.f.MinSize()
}

func (r *paneFixedHeightRenderer) Refresh() {}

func (r *paneFixedHeightRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.child}
}

func (r *paneFixedHeightRenderer) Destroy() {}

func (f *paneFixedHeight) CreateRenderer() fyne.WidgetRenderer {
	return &paneFixedHeightRenderer{f: f, child: f.child}
}

type paneIconButton struct {
	widget.BaseWidget
	icon    fyne.Resource
	onTap   func()
	hovered bool
}

func newPaneIconButton(icon fyne.Resource, onTap func()) *paneIconButton {
	b := &paneIconButton{icon: icon, onTap: onTap}
	b.ExtendBaseWidget(b)
	return b
}

func (b *paneIconButton) MinSize() fyne.Size {
	return fyne.NewSize(paneIconBtnSize, paneIconBtnSize)
}

func (b *paneIconButton) Tapped(*fyne.PointEvent) {
	if b.onTap != nil {
		b.onTap()
	}
}

func (b *paneIconButton) MouseIn(*desktop.MouseEvent) {
	b.hovered = true
	b.Refresh()
}

func (b *paneIconButton) MouseMoved(*desktop.MouseEvent) {}

func (b *paneIconButton) MouseOut() {
	b.hovered = false
	b.Refresh()
}

func (b *paneIconButton) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

type paneIconButtonRenderer struct {
	btn  *paneIconButton
	bg   *canvas.Rectangle
	icon *canvas.Image
}

func (r *paneIconButtonRenderer) Layout(size fyne.Size) {
	r.btn.Resize(size)
	r.bg.Resize(size)
	r.bg.Move(fyne.NewPos(0, 0))
	iconSz := fyne.NewSize(paneIconGlyphSize, paneIconGlyphSize)
	r.icon.Resize(iconSz)
	r.icon.Move(fyne.NewPos((size.Width-iconSz.Width)/2, (size.Height-iconSz.Height)/2))
}

func (r *paneIconButtonRenderer) MinSize() fyne.Size {
	return r.btn.MinSize()
}

func (r *paneIconButtonRenderer) Refresh() {
	if r.btn.hovered {
		r.bg.FillColor = colorRowHover
	} else {
		r.bg.FillColor = color.Transparent
	}
	canvas.Refresh(r.bg)
}

func (r *paneIconButtonRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.icon}
}

func (r *paneIconButtonRenderer) Destroy() {}

func (b *paneIconButton) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	img := canvas.NewImageFromResource(b.icon)
	img.FillMode = canvas.ImageFillContain
	return &paneIconButtonRenderer{btn: b, bg: bg, icon: img}
}

var _ fyne.Tappable = (*paneIconButton)(nil)
var _ desktop.Hoverable = (*paneIconButton)(nil)
var _ desktop.Cursorable = (*paneIconButton)(nil)

type paneTapLabel struct {
	widget.BaseWidget
	text  *canvas.Text
	onTap func()
}

func newPaneTapLabel(text string, c color.Color, size float32, onTap func()) *paneTapLabel {
	t := canvas.NewText(text, c)
	t.TextSize = size
	l := &paneTapLabel{text: t, onTap: onTap}
	l.ExtendBaseWidget(l)
	return l
}

func (l *paneTapLabel) SetText(text string) {
	l.text.Text = text
	canvas.Refresh(l.text)
	l.Refresh()
}

func (l *paneTapLabel) MinSize() fyne.Size {
	sz, _ := fyne.CurrentApp().Driver().RenderedTextSize(l.text.Text, l.text.TextSize, l.text.TextStyle, l.text.FontSource)
	if sz.Height < l.text.TextSize {
		sz.Height = l.text.TextSize
	}
	if sz.Width < 1 {
		sz.Width = 1
	}
	return sz
}

func (l *paneTapLabel) Tapped(*fyne.PointEvent) {
	if l.onTap != nil {
		l.onTap()
	}
}

func (l *paneTapLabel) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

type paneTapLabelRenderer struct {
	label *paneTapLabel
	text  *canvas.Text
}

func (r *paneTapLabelRenderer) Layout(size fyne.Size) {
	r.label.Resize(size)
	r.text.Resize(size)
	r.text.Move(fyne.NewPos(0, (size.Height-r.text.MinSize().Height)/2))
}

func (r *paneTapLabelRenderer) MinSize() fyne.Size {
	return r.label.MinSize()
}

func (r *paneTapLabelRenderer) Refresh() {
	canvas.Refresh(r.text)
}

func (r *paneTapLabelRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.text}
}

func (r *paneTapLabelRenderer) Destroy() {}

func (l *paneTapLabel) CreateRenderer() fyne.WidgetRenderer {
	return &paneTapLabelRenderer{label: l, text: l.text}
}

var _ fyne.Tappable = (*paneTapLabel)(nil)
var _ desktop.Cursorable = (*paneTapLabel)(nil)

func paneChromeEntry(entry *widget.Entry) fyne.CanvasObject {
	themed := container.NewThemeOverride(entry, newPanePathEntryTheme(paneRowNameSize))
	return newPaneFixedHeight(paneBandInnerHeight(), themed)
}

func paneChromeVSeparator() fyne.CanvasObject {
	line := canvas.NewRectangle(colorBorder)
	line.SetMinSize(fyne.NewSize(1, paneBandInnerHeight()-6))
	return container.NewCenter(line)
}
