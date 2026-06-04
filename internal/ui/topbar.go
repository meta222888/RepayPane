package ui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type TopBar struct {
	app        *App
	root       fyne.CanvasObject
	btnSet     *widget.Button
	btnFeat    *widget.Button
	btnAbout   *widget.Button
	maximized  bool
}

func NewTopBar(app *App) *TopBar {
	t := &TopBar{app: app}
	logo := widget.NewLabelWithStyle("⇄  "+i18n.T(i18n.KeyAppTitle), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	t.btnSet = widget.NewButtonWithIcon(i18n.T(i18n.KeyMenuSettings), theme.SettingsIcon(), func() {
		t.showMenu(app.settingsMenu(), t.btnSet)
	})
	t.btnFeat = widget.NewButtonWithIcon(i18n.T(i18n.KeyMenuFeatures), theme.ViewRefreshIcon(), func() {
		t.showMenu(app.featuresMenu(), t.btnFeat)
	})
	t.btnAbout = widget.NewButtonWithIcon(i18n.T(i18n.KeyMenuAbout), theme.InfoIcon(), func() {
		t.showMenu(app.aboutMenu(), t.btnAbout)
	})
	t.btnSet.Importance = widget.LowImportance
	t.btnFeat.Importance = widget.LowImportance
	t.btnAbout.Importance = widget.LowImportance

	minBtn := widget.NewButtonWithIcon("", theme.WindowMinimizeIcon(), func() {
		minimizeWindow(t.app.window)
	})
	maxBtn := widget.NewButtonWithIcon("", theme.WindowMaximizeIcon(), func() {
		toggleMaximizeWindow(t.app.window, &t.maximized)
	})
	closeBtn := widget.NewButtonWithIcon("", theme.WindowCloseIcon(), func() {
		closeWindow(t.app.window)
	})
	for _, b := range []*widget.Button{minBtn, maxBtn, closeBtn} {
		b.Importance = widget.LowImportance
	}
	winControls := container.NewHBox(minBtn, maxBtn, closeBtn)

	logoDrag := newDragRegion(t.app.window, logo)
	right := container.NewHBox(t.btnSet, t.btnFeat, t.btnAbout, winControls)
	barContent := container.NewBorder(nil, nil, logoDrag, right, nil)
	dragLayer := newDragRegion(t.app.window, layout.NewSpacer())
	header := container.NewStack(dragLayer, barContent)
	wrapped := withPanelHeader(header)
	t.root = wrapped
	return t
}

func (t *TopBar) Container() fyne.CanvasObject { return t.root }

func (t *TopBar) ApplyLanguage() {
	t.btnSet.SetText(i18n.T(i18n.KeyMenuSettings))
	t.btnFeat.SetText(i18n.T(i18n.KeyMenuFeatures))
	t.btnAbout.SetText(i18n.T(i18n.KeyMenuAbout))
}

func (t *TopBar) showMenu(m *fyne.Menu, rel fyne.CanvasObject) {
	c := t.app.window.Canvas()
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(rel)
	pop := widget.NewPopUpMenu(m, c)
	pop.ShowAtPosition(pos.Add(fyne.NewPos(0, rel.MinSize().Height)))
}

func (a *App) settingsMenu() *fyne.Menu {
	langItem := fyne.NewMenuItem(i18n.T(i18n.KeyMenuLanguage), nil)
	langItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem(i18n.T(i18n.KeyMenuLangEN), func() { a.setLanguage(i18n.EN) }),
		fyne.NewMenuItem(i18n.T(i18n.KeyMenuLangZH), func() { a.setLanguage(i18n.ZH) }),
	)
	return fyne.NewMenu("",
		langItem,
		fyne.NewMenuItem(i18n.T(i18n.KeyMenuMyServers), a.showMyServers),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(i18n.T(i18n.KeyMenuExit), func() { a.fyneApp.Quit() }),
	)
}

func (a *App) featuresMenu() *fyne.Menu {
	return fyne.NewMenu("",
		fyne.NewMenuItem(i18n.T(i18n.KeyMenuComingSoon), a.showFeaturesSoon),
	)
}

func (a *App) aboutMenu() *fyne.Menu {
	return fyne.NewMenu("",
		fyne.NewMenuItem(i18n.T(i18n.KeyMenuCheckUpdate), a.showCheckUpdate),
		fyne.NewMenuItem(i18n.T(i18n.KeyMenuAboutUs), a.showAboutUs),
	)
}
