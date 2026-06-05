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
	serverPickerIcon     = "🖥"
	pickerRowNameSize    float32 = 14
	pickerRowMinHeight   float32 = 30
)

type connectPickerRow struct {
	widget.BaseWidget

	rowIndex int
	selected bool
	hovered  bool

	bg    *canvas.Rectangle
	iconT *canvas.Text
	lineT *canvas.Text

	onPrimary func()
	onDouble  func()
}

func newConnectPickerRow() *connectPickerRow {
	r := &connectPickerRow{}
	r.ExtendBaseWidget(r)
	return r
}

func (r *connectPickerRow) update(rowIndex int, icon, name, subtitle string, selected bool) {
	r.rowIndex = rowIndex
	r.selected = selected
	if r.lineT == nil {
		return
	}
	if icon == "" {
		icon = serverPickerIcon
	}
	r.iconT.Text = icon
	if subtitle != "" {
		r.lineT.Text = name + "  ·  " + subtitle
	} else {
		r.lineT.Text = name
	}
	r.refreshStyle()
}

func (r *connectPickerRow) rowBgColor() color.Color {
	if r.selected {
		return colorRowSelected
	}
	if r.hovered {
		return colorRowHover
	}
	if r.rowIndex%2 == 0 {
		return colorPanel
	}
	return colorRowAlt
}

func (r *connectPickerRow) refreshStyle() {
	if r.bg == nil {
		return
	}
	r.bg.FillColor = r.rowBgColor()
	if r.selected {
		r.lineT.Color = colorTextHighlight
	} else {
		r.lineT.Color = colorForeground
	}
	canvas.Refresh(r.bg)
	canvas.Refresh(r.iconT)
	canvas.Refresh(r.lineT)
}

func (r *connectPickerRow) Tapped(*fyne.PointEvent) {
	if r.onPrimary != nil {
		r.onPrimary()
	}
}

func (r *connectPickerRow) DoubleTapped(*fyne.PointEvent) {
	if r.onDouble != nil {
		r.onDouble()
	}
}

func (r *connectPickerRow) MouseIn(_ *desktop.MouseEvent) {
	r.hovered = true
	r.refreshStyle()
}

func (r *connectPickerRow) MouseMoved(_ *desktop.MouseEvent) {}

func (r *connectPickerRow) MouseOut() {
	r.hovered = false
	r.refreshStyle()
}

func (r *connectPickerRow) CreateRenderer() fyne.WidgetRenderer {
	r.bg = canvas.NewRectangle(r.rowBgColor())
	r.bg.SetMinSize(fyne.NewSize(0, pickerRowMinHeight))
	r.iconT = canvas.NewText(serverPickerIcon, colorAccent)
	r.iconT.TextSize = 11
	r.lineT = canvas.NewText("", colorForeground)
	r.lineT.TextSize = pickerRowNameSize

	nameBox := container.NewHBox(r.iconT, r.lineT)
	content := container.NewStack(r.bg, container.NewPadded(nameBox))
	return widget.NewSimpleRenderer(content)
}

func (r *connectPickerRow) MinSize() fyne.Size {
	return fyne.NewSize(0, pickerRowMinHeight)
}

var _ fyne.Tappable = (*connectPickerRow)(nil)
var _ fyne.DoubleTappable = (*connectPickerRow)(nil)
var _ desktop.Hoverable = (*connectPickerRow)(nil)
