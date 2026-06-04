package ui

import (
	"net/url"
	"os"
	"strings"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/version"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (a *App) setLanguage(lang i18n.Lang) {
	if i18n.Current() == lang {
		return
	}
	i18n.SetLanguage(lang)
	a.settings.Language = string(lang)
	_ = config.SaveSettings(a.settings)
	a.applyLanguage()
}

func (a *App) showFeaturesSoon() {
	dialog.ShowInformation(i18n.T(i18n.KeyMenuFeatures), i18n.T(i18n.KeyMenuFeaturesSoon), a.window)
}

func (a *App) showCheckUpdate() {
	dialog.ShowInformation(i18n.T(i18n.KeyCheckUpdateTitle), i18n.T(i18n.KeyCheckUpdateMsg), a.window)
}

func (a *App) showAboutUs() {
	intro := widget.NewLabel(i18n.T(i18n.KeyAboutIntro))
	intro.Wrapping = fyne.TextWrapWord
	versionLabel := widget.NewLabel(i18n.Tf(i18n.KeyAboutVersion, version.Version))
	site, _ := url.Parse(i18n.T(i18n.KeyAboutWebsite))
	website := widget.NewHyperlink(i18n.T(i18n.KeyAboutWebsite), site)

	content := container.NewVBox(intro, versionLabel, website)
	d := dialog.NewCustom(i18n.T(i18n.KeyAboutTitle), i18n.T(i18n.KeyOK), content, a.window)
	d.Resize(fyne.NewSize(420, 200))
	d.Show()
}

func (a *App) showMyServers() {
	list := widget.NewList(
		func() int { return len(a.store.Servers) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			s := a.store.Servers[id]
			obj.(*widget.Label).SetText(s.Name + "  (" + s.Host + ")")
		},
	)
	selected := -1
	list.OnSelected = func(id widget.ListItemID) { selected = int(id) }

	buttons := container.NewHBox(
		widget.NewButton(i18n.T(i18n.KeyAddServer), func() {
			a.onNewTab()
			list.Refresh()
		}),
		widget.NewButton(i18n.T(i18n.KeyEdit), func() {
			if selected < 0 {
				dialogShow(a, i18n.T(i18n.KeySelectServer), i18n.T(i18n.KeyChooseEdit))
				return
			}
			a.selectedServerID = selected
			a.showEditServer()
			list.Refresh()
		}),
		widget.NewButton(i18n.T(i18n.KeyDelete), func() {
			if selected < 0 {
				dialogShow(a, i18n.T(i18n.KeySelectServer), i18n.T(i18n.KeyChooseDelete))
				return
			}
			a.selectedServerID = selected
			a.showDeleteServer()
			list.Refresh()
		}),
	)

	content := container.NewBorder(nil, buttons, nil, nil, list)
	d := dialog.NewCustom(i18n.T(i18n.KeyMyServersTitle), i18n.T(i18n.KeyOK), content, a.window)
	d.Resize(fyne.NewSize(480, 360))
	d.Show()
}

func initLanguage(settings *config.Settings) {
	if settings.Language == "zh" {
		i18n.SetLanguage(i18n.ZH)
		return
	}
	if settings.Language == "en" {
		i18n.SetLanguage(i18n.EN)
		return
	}
	if lang := os.Getenv("LANG"); strings.HasPrefix(strings.ToLower(lang), "en") {
		i18n.SetLanguage(i18n.EN)
		settings.Language = "en"
		return
	}
	i18n.SetLanguage(i18n.ZH)
	settings.Language = "zh"
}
