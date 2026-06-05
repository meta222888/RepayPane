package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const paneFileIconGap float32 = 4

// paneFileNameCell shows a file row icon and ellipsized name on one baseline-centered row.
type paneFileNameCell struct {
	widget.BaseWidget
	textSize float32
	icon     *canvas.Text
	name     *paneEllipsisText
}

func newPaneFileNameCell(textSize float32, c color.Color) *paneFileNameCell {
	p := &paneFileNameCell{textSize: textSize}
	p.ExtendBaseWidget(p)
	return p
}

func (p *paneFileNameCell) SetContent(icon, name string) {
	if p.icon == nil {
		return
	}
	p.icon.Text = icon
	p.name.SetText(name)
	p.Refresh()
}

func (p *paneFileNameCell) SetColor(c color.Color) {
	if p.icon == nil {
		return
	}
	p.icon.Color = c
	p.name.SetColor(c)
}

func (p *paneFileNameCell) Hide() {
	if p.icon == nil {
		return
	}
	p.icon.Hide()
	p.name.Hide()
	canvas.Refresh(p.icon)
	canvas.Refresh(p.name)
}

func (p *paneFileNameCell) Show() {
	if p.icon == nil {
		return
	}
	p.icon.Show()
	p.name.Show()
	canvas.Refresh(p.icon)
	canvas.Refresh(p.name)
}

func (p *paneFileNameCell) MinSize() fyne.Size {
	return fyne.NewSize(0, p.textSize)
}

type paneFileNameCellRenderer struct {
	cell *paneFileNameCell
	icon *canvas.Text
	name *paneEllipsisText
}

func (r *paneFileNameCellRenderer) Layout(size fyne.Size) {
	r.cell.Resize(size)
	driver := fyne.CurrentApp().Driver()

	iconSz, _ := driver.RenderedTextSize(r.icon.Text, r.icon.TextSize, r.icon.TextStyle, r.icon.FontSource)
	if iconSz.Width < 1 {
		iconSz.Width = 1
	}
	if iconSz.Height < r.icon.TextSize {
		iconSz.Height = r.icon.TextSize
	}

	nameX := float32(0)
	if r.icon.Text != "" {
		nameX = iconSz.Width + paneFileIconGap
		r.icon.Move(fyne.NewPos(0, paneRowTextCenterY(size.Height, iconSz.Height)))
		r.icon.Show()
	} else {
		r.icon.Hide()
	}

	nameW := size.Width - nameX
	if nameW < 0 {
		nameW = 0
	}
	r.name.Resize(fyne.NewSize(nameW, size.Height))
	r.name.Move(fyne.NewPos(nameX, 0))
}

func (r *paneFileNameCellRenderer) MinSize() fyne.Size {
	return r.cell.MinSize()
}

func (r *paneFileNameCellRenderer) Refresh() {
	canvas.Refresh(r.icon)
	r.name.Refresh()
}

func (r *paneFileNameCellRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.icon, r.name}
}

func (r *paneFileNameCellRenderer) Destroy() {}

func (p *paneFileNameCell) CreateRenderer() fyne.WidgetRenderer {
	p.icon = canvas.NewText("", colorForeground)
	p.icon.TextSize = p.textSize
	p.name = newPaneEllipsisText(p.textSize, colorForeground)
	return &paneFileNameCellRenderer{cell: p, icon: p.icon, name: p.name}
}
