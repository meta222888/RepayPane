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
	title := i18n.T(i18n.KeyAboutTitle)
	var dlg *modalDialog

	intro := widget.NewLabel(i18n.T(i18n.KeyAboutIntro))
	intro.Wrapping = fyne.TextWrapWord
	versionLabel := widget.NewLabel(i18n.Tf(i18n.KeyAboutVersion, version.Version))
	site, _ := url.Parse(i18n.T(i18n.KeyAboutWebsite))
	website := widget.NewHyperlink(i18n.T(i18n.KeyAboutWebsite), site)

	content := container.NewVBox(intro, versionLabel, website)
	closeBtn := newAccentButton(i18n.T(i18n.KeyOK), func() { dlg.Close() })
	body := container.NewBorder(nil, container.NewHBox(closeBtn), nil, nil, content)
	dlg = newModalDialog(a.window, title, fyne.NewSize(420, 220), body)
}

func (a *App) showMyServers() {
	title := i18n.T(i18n.KeyMyServersTitle)
	var dlg *modalDialog

	selected := -1
	prevSelected := -1
	list := widget.NewList(
		func() int { return len(a.store.Servers) },
		func() fyne.CanvasObject { return newConnectPickerRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if int(id) >= len(a.store.Servers) {
				return
			}
			s := a.store.Servers[id]
			name := s.Name
			if name == "" {
				name = s.Host
			}
			obj.(*connectPickerRow).update(name, s.Host, int(id) == selected)
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		if int(id) >= len(a.store.Servers) {
			return
		}
		prevSelected = selected
		selected = int(id)
		if prevSelected >= 0 {
			list.RefreshItem(widget.ListItemID(prevSelected))
		}
		list.RefreshItem(id)
	}

	buttons := container.NewHBox(
		newAccentButton(i18n.T(i18n.KeyAddServer), func() {
			dlg.Close()
			a.onNewTab()
		}),
		newAccentButton(i18n.T(i18n.KeyEdit), func() {
			if selected < 0 {
				dialogShow(a, i18n.T(i18n.KeySelectServer), i18n.T(i18n.KeyChooseEdit))
				return
			}
			a.selectedServerID = selected
			dlg.Close()
			a.showEditServer()
		}),
		newAccentButton(i18n.T(i18n.KeyDelete), func() {
			if selected < 0 {
				dialogShow(a, i18n.T(i18n.KeySelectServer), i18n.T(i18n.KeyChooseDelete))
				return
			}
			a.selectedServerID = selected
			dlg.Close()
			a.showDeleteServer()
		}),
		newAccentButton(i18n.T(i18n.KeyOK), func() { dlg.Close() }),
	)

	body := container.NewBorder(nil, buttons, nil, nil, list)
	dlg = newModalDialog(a.window, title, fyne.NewSize(480, 360), body)
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
