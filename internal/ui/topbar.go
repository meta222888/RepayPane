package ui

import (
	"github.com/relaypane/relaypane/internal/assets"
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type TopBar struct {
	app     *App
	root    fyne.CanvasObject
	btnSet  *widget.Button
	btnFeat *widget.Button
	btnAbout *widget.Button
}

func NewTopBar(app *App) *TopBar {
	t := &TopBar{app: app}
	logoImg := canvas.NewImageFromResource(assets.LogoResource())
	logoImg.FillMode = canvas.ImageFillContain
	logoImg.SetMinSize(fyne.NewSize(22, 22))
	appName := canvas.NewText(i18n.T(i18n.KeyAppTitle), colorForeground)
	appName.TextSize = 14
	appName.TextStyle = fyne.TextStyle{Bold: true}
	logo := container.NewHBox(logoImg, wrapCanvasText(appName))
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

	right := container.NewHBox(t.btnSet, t.btnFeat, t.btnAbout)
	barContent := container.NewBorder(nil, nil, logo, right, layout.NewSpacer())
	wrapped := withPanelHeader(barContent)
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
		fyne.NewMenuItem(i18n.T(i18n.KeyMenuCloudSync), a.showCloudSync),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(i18n.T(i18n.KeyMenuExit), a.prepareQuit),
	)
}

func (a *App) featuresMenu() *fyne.Menu {
	syncItem := fyne.NewMenuItem(i18n.T(i18n.KeyFeatSync), nil)
	syncItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem(i18n.T(i18n.KeyFeatSyncUp), a.syncLocalToRemote),
		fyne.NewMenuItem(i18n.T(i18n.KeyFeatSyncDown), a.syncRemoteToLocal),
	)
	return fyne.NewMenu("",
		fyne.NewMenuItem(i18n.T(i18n.KeyFeatSysInfo), a.showSystemInfo),
		fyne.NewMenuItem(i18n.T(i18n.KeyFeatNetwork), a.showNetworkInfo),
		fyne.NewMenuItem(i18n.T(i18n.KeyFeatDisk), a.showDiskSpace),
		fyne.NewMenuItem(i18n.T(i18n.KeyFeatDu), a.showDiskUsageTree),
		fyne.NewMenuItem(i18n.T(i18n.KeyFeatResources), a.showResourceUsage),
		syncItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(i18n.T(i18n.KeyFeatShell), a.showRemoteShell),
	)
}

func (a *App) aboutMenu() *fyne.Menu {
	return fyne.NewMenu("",
		fyne.NewMenuItem(i18n.T(i18n.KeyMenuCheckUpdate), a.showCheckUpdate),
		fyne.NewMenuItem(i18n.T(i18n.KeyMenuAboutUs), a.showAboutUs),
	)
}
