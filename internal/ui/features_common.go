package ui

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) requireClient() (*remote.Client, bool) {
	c := a.activeClient()
	if c == nil {
		dialogShow(a, i18n.T(i18n.KeyNotConnectedTitle), i18n.T(i18n.KeyNotConnectedFirst))
		return nil, false
	}
	return c, true
}

func runRemoteAsync(a *App, fn func(*remote.Client) (string, error), setText func(string), setErr func(error)) {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	setText(i18n.T(i18n.KeyFeatLoading))
	go func() {
		out, err := fn(client)
		fyne.Do(func() {
			if err != nil {
				setErr(err)
				return
			}
			setText(out)
		})
	}()
}

func monoOutput(text string) fyne.CanvasObject {
	lbl := widget.NewLabel(text)
	lbl.Wrapping = fyne.TextWrapWord
	return container.NewScroll(lbl)
}

func scrollLabel() (*widget.Label, fyne.CanvasObject) {
	lbl := widget.NewLabel("")
	lbl.Wrapping = fyne.TextWrapWord
	return lbl, container.NewScroll(lbl)
}

type usageProgressBar struct {
	widget.BaseWidget
	value     float64
	fillColor color.Color
}

func newUsageProgressBar() *usageProgressBar {
	p := &usageProgressBar{fillColor: colorConnected}
	p.ExtendBaseWidget(p)
	return p
}

func (p *usageProgressBar) SetUsage(pct float64) {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	p.value = pct / 100
	switch {
	case pct >= 90:
		p.fillColor = color.NRGBA{R: 220, G: 90, B: 90, A: 255}
	case pct >= 70:
		p.fillColor = colorWarning
	default:
		p.fillColor = colorConnected
	}
	p.Refresh()
}

func (p *usageProgressBar) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(colorInput)
	fill := canvas.NewRectangle(p.fillColor)
	return &usageProgressRenderer{bar: p, bg: bg, fill: fill}
}

type usageProgressRenderer struct {
	bar  *usageProgressBar
	bg   *canvas.Rectangle
	fill *canvas.Rectangle
}

func (r *usageProgressRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.bg.Move(fyne.NewPos(0, 0))
	fillW := size.Width * float32(r.bar.value)
	r.fill.Resize(fyne.NewSize(fillW, size.Height))
	r.fill.Move(fyne.NewPos(0, 0))
}

func (r *usageProgressRenderer) MinSize() fyne.Size {
	return fyne.NewSize(120, 10)
}

func (r *usageProgressRenderer) Refresh() {
	r.fill.FillColor = r.bar.fillColor
	canvas.Refresh(r.bg)
	canvas.Refresh(r.fill)
}

func (r *usageProgressRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.fill}
}

func (r *usageProgressRenderer) Destroy() {}

func diskUsageCard(mount, total, used, avail, pctStr string) fyne.CanvasObject {
	pct, _ := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(pctStr), "%"), 64)
	title := widget.NewLabelWithStyle(mount, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	detail := widget.NewLabel(i18n.Tf(i18n.KeyFeatDiskDetail, used, total, avail, pctStr))
	bar := newUsageProgressBar()
	bar.SetUsage(pct)
	row := container.NewVBox(title, bar, detail)
	return withBackground(container.NewPadded(row), colorPanel)
}

func showThemedFeature(a *App, title string, size fyne.Size, body fyne.CanvasObject) *modalDialog {
	return newModalDialog(a.window, title, size, body)
}

func confirmThemed(a *App, title, msg string, onOK func()) {
	var dlg *modalDialog
	ok := newAccentButton(i18n.T(i18n.KeyOK), func() {
		dlg.Close()
		onOK()
	})
	cancel := newAccentButton(i18n.T(i18n.KeyCancel), func() { dlg.Close() })
	btns := container.NewHBox(cancel, ok)
	lbl := widget.NewLabel(msg)
	lbl.Wrapping = fyne.TextWrapWord
	body := container.NewBorder(nil, btns, nil, nil, lbl)
	dlg = newModalDialog(a.window, title, fyne.NewSize(480, 220), body)
}
