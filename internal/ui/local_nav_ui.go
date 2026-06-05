package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	localNavPopupWidth   float32 = 240
	localNavMenuRowH     float32 = 30
	localNavDriveBtnMinW float32 = 58
	localNavIconSize     float32 = 14
)

type localDriveButton struct {
	widget.BaseWidget
	label   string
	hovered bool
	onTap   func()
}

func newLocalDriveButton(onTap func()) *localDriveButton {
	b := &localDriveButton{onTap: onTap}
	b.ExtendBaseWidget(b)
	return b
}

func (b *localDriveButton) SetLabel(text string) {
	if b.label == text {
		return
	}
	b.label = text
	b.Refresh()
}

func (b *localDriveButton) MinSize() fyne.Size {
	return fyne.NewSize(localNavDriveBtnMinW, paneBandInnerHeight())
}

func (b *localDriveButton) Tapped(*fyne.PointEvent) {
	if b.onTap != nil {
		b.onTap()
	}
}

func (b *localDriveButton) MouseIn(*desktop.MouseEvent) {
	b.hovered = true
	b.Refresh()
}

func (b *localDriveButton) MouseMoved(*desktop.MouseEvent) {}

func (b *localDriveButton) MouseOut() {
	b.hovered = false
	b.Refresh()
}

func (b *localDriveButton) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

type localDriveButtonRenderer struct {
	btn     *localDriveButton
	border  *canvas.Rectangle
	bg      *canvas.Rectangle
	icon    *canvas.Image
	label   *canvas.Text
	chevron *canvas.Text
}

func (r *localDriveButtonRenderer) Layout(size fyne.Size) {
	r.btn.Resize(size)
	r.border.Resize(size)
	r.border.Move(fyne.NewPos(0, 0))

	inset := float32(1)
	inner := fyne.NewSize(size.Width-inset*2, size.Height-inset*2)
	r.bg.Resize(inner)
	r.bg.Move(fyne.NewPos(inset, inset))

	pad := float32(4)
	x := pad
	iconSz := fyne.NewSize(localNavIconSize, localNavIconSize)
	r.icon.Resize(iconSz)
	r.icon.Move(fyne.NewPos(x+(localNavIconSize-iconSz.Width)/2, (size.Height-iconSz.Height)/2))
	x += localNavIconSize + 3

	lblSz, _ := fyne.CurrentApp().Driver().RenderedTextSize(r.label.Text, r.label.TextSize, r.label.TextStyle, r.label.FontSource)
	r.label.Move(fyne.NewPos(x, (size.Height-lblSz.Height)/2))

	chevSz, _ := fyne.CurrentApp().Driver().RenderedTextSize(r.chevron.Text, r.chevron.TextSize, r.chevron.TextStyle, r.chevron.FontSource)
	r.chevron.Move(fyne.NewPos(size.Width-pad-chevSz.Width, (size.Height-chevSz.Height)/2))
}

func (r *localDriveButtonRenderer) MinSize() fyne.Size {
	return r.btn.MinSize()
}

func (r *localDriveButtonRenderer) Refresh() {
	r.label.Text = r.btn.label
	if r.btn.hovered {
		r.bg.FillColor = colorRowHover
	} else {
		r.bg.FillColor = colorInput
	}
	r.border.FillColor = colorBorder
	canvas.Refresh(r.border)
	canvas.Refresh(r.bg)
	canvas.Refresh(r.icon)
	canvas.Refresh(r.label)
	canvas.Refresh(r.chevron)
}

func (r *localDriveButtonRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.border, r.bg, r.icon, r.label, r.chevron}
}

func (r *localDriveButtonRenderer) Destroy() {}

func (b *localDriveButton) CreateRenderer() fyne.WidgetRenderer {
	icon := canvas.NewImageFromResource(theme.ComputerIcon())
	icon.FillMode = canvas.ImageFillContain
	lbl := canvas.NewText(b.label, colorForeground)
	lbl.TextSize = paneRowNameSize
	chev := canvas.NewText("▾", colorMuted)
	chev.TextSize = AppTextSize
	return &localDriveButtonRenderer{
		btn:     b,
		border:  canvas.NewRectangle(colorBorder),
		bg:      canvas.NewRectangle(colorInput),
		icon:    icon,
		label:   lbl,
		chevron: chev,
	}
}

var _ fyne.Tappable = (*localDriveButton)(nil)
var _ desktop.Hoverable = (*localDriveButton)(nil)
var _ desktop.Cursorable = (*localDriveButton)(nil)

type localNavMenuRow struct {
	widget.BaseWidget
	iconRes  fyne.Resource
	title    string
	subtitle string
	accent   bool
	hovered  bool
	onTap    func()
}

func newLocalNavMenuRow(icon fyne.Resource, title, subtitle string, accent bool, onTap func()) *localNavMenuRow {
	r := &localNavMenuRow{iconRes: icon, title: title, subtitle: subtitle, accent: accent, onTap: onTap}
	r.ExtendBaseWidget(r)
	return r
}

func (r *localNavMenuRow) MinSize() fyne.Size {
	return fyne.NewSize(localNavPopupWidth-16, localNavMenuRowH)
}

func (r *localNavMenuRow) Tapped(*fyne.PointEvent) {
	if r.onTap != nil {
		r.onTap()
	}
}

func (r *localNavMenuRow) MouseIn(*desktop.MouseEvent) {
	r.hovered = true
	r.Refresh()
}

func (r *localNavMenuRow) MouseMoved(*desktop.MouseEvent) {}

func (r *localNavMenuRow) MouseOut() {
	r.hovered = false
	r.Refresh()
}

func (r *localNavMenuRow) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

type localNavMenuRowRenderer struct {
	row      *localNavMenuRow
	bg       *canvas.Rectangle
	icon     *canvas.Image
	titleT   *canvas.Text
	subtitle *canvas.Text
}

func (r *localNavMenuRowRenderer) Layout(size fyne.Size) {
	r.row.Resize(size)
	r.bg.Resize(size)
	r.bg.Move(fyne.NewPos(0, 0))

	pad := float32(10)
	iconSz := fyne.NewSize(localNavIconSize, localNavIconSize)
	r.icon.Resize(iconSz)
	r.icon.Move(fyne.NewPos(pad, (size.Height-iconSz.Height)/2))

	titleX := pad + localNavIconSize + 8
	titleSz, _ := fyne.CurrentApp().Driver().RenderedTextSize(r.titleT.Text, r.titleT.TextSize, r.titleT.TextStyle, r.titleT.FontSource)
	r.titleT.Move(fyne.NewPos(titleX, (size.Height-titleSz.Height)/2))

	if r.subtitle.Text != "" {
		subSz, _ := fyne.CurrentApp().Driver().RenderedTextSize(r.subtitle.Text, r.subtitle.TextSize, r.subtitle.TextStyle, r.subtitle.FontSource)
		r.subtitle.Move(fyne.NewPos(size.Width-pad-subSz.Width, (size.Height-subSz.Height)/2))
	}
}

func (r *localNavMenuRowRenderer) MinSize() fyne.Size {
	return r.row.MinSize()
}

func (r *localNavMenuRowRenderer) Refresh() {
	r.titleT.Text = r.row.title
	r.subtitle.Text = r.row.subtitle
	if r.row.accent {
		r.titleT.Color = colorAccent
	} else if r.row.hovered {
		r.titleT.Color = colorTextHighlight
	} else {
		r.titleT.Color = colorForeground
	}
	if r.row.hovered {
		r.bg.FillColor = colorRowHover
	} else {
		r.bg.FillColor = color.Transparent
	}
	canvas.Refresh(r.bg)
	canvas.Refresh(r.icon)
	canvas.Refresh(r.titleT)
	canvas.Refresh(r.subtitle)
}

func (r *localNavMenuRowRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.icon, r.titleT, r.subtitle}
}

func (r *localNavMenuRowRenderer) Destroy() {}

func (r *localNavMenuRow) CreateRenderer() fyne.WidgetRenderer {
	icon := canvas.NewImageFromResource(r.iconRes)
	icon.FillMode = canvas.ImageFillContain
	title := canvas.NewText(r.title, colorForeground)
	title.TextSize = paneRowNameSize
	sub := canvas.NewText(r.subtitle, colorMuted)
	sub.TextSize = paneRowMetaSize
	return &localNavMenuRowRenderer{row: r, bg: canvas.NewRectangle(color.Transparent), icon: icon, titleT: title, subtitle: sub}
}

var _ fyne.Tappable = (*localNavMenuRow)(nil)
var _ desktop.Hoverable = (*localNavMenuRow)(nil)
var _ desktop.Cursorable = (*localNavMenuRow)(nil)

func localNavSectionHeader(text string) fyne.CanvasObject {
	lbl := canvas.NewText(text, colorMuted)
	lbl.TextSize = AppTitleTextSize
	return container.New(layout.NewCustomPaddedLayout(10, 2, 10, 0), lbl)
}

func localNavPopupSeparator() fyne.CanvasObject {
	line := canvas.NewRectangle(colorBorder)
	line.SetMinSize(fyne.NewSize(0, 1))
	return container.New(layout.NewCustomPaddedLayout(8, 6, 8, 6), line)
}

func newLocalNavPopupPanel(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(colorPanel)
	bg.CornerRadius = 4
	padded := container.New(layout.NewCustomPaddedLayout(4, 6, 4, 6), content)
	panel := container.NewStack(bg, padded)
	minW := canvas.NewRectangle(color.Transparent)
	minW.SetMinSize(fyne.NewSize(localNavPopupWidth, 0))
	return container.NewStack(minW, panel)
}

func showLocalNavPopup(c fyne.Canvas, anchor fyne.CanvasObject, build func(dismiss func()) fyne.CanvasObject) {
	dismissPopUpMenus(c)
	content := build(func() {
		for _, o := range c.Overlays().List() {
			o.Hide()
		}
	})
	panel := newLocalNavPopupPanel(content)
	size := panel.MinSize()
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(anchor)
	at := pos.Add(fyne.NewPos(0, anchor.MinSize().Height))
	at = adjustMenuPosition(at, size, c)
	pop := widget.NewPopUp(panel, c)
	pop.Resize(size)
	pop.ShowAtPosition(at)
}
