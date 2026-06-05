package ui

import (
	"net/url"
	"time"

	"github.com/relaypane/relaypane/internal/cloudsync"
	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/fileopen"
	"github.com/relaypane/relaypane/internal/i18n"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (a *App) showCloudSync() {
	var cloudDlg *modalDialog
	parent := func() fyne.Window {
		if cloudDlg != nil && cloudDlg.Window() != nil {
			return cloudDlg.Window()
		}
		return a.window
	}

	apiSecret := widget.NewPasswordEntry()
	apiSecret.SetText(a.settings.CloudSyncAPISecret)
	apiSecret.SetPlaceHolder(i18n.T(i18n.KeyCloudSyncAPISecretHint))

	apiSecretURL, _ := url.Parse("https://pc530.com/easystorage/")
	apiSecretLink := widget.NewHyperlink(i18n.T(i18n.KeyCloudSyncAPISecretLink), apiSecretURL)

	encPass := widget.NewPasswordEntry()
	encPass.SetText(a.settings.CloudSyncPassword)
	encPass.SetPlaceHolder(i18n.T(i18n.KeyCloudSyncPasswordHint))

	privacyNote := widget.NewLabel(i18n.T(i18n.KeyCloudSyncPrivacyNote))
	privacyNote.Wrapping = fyne.TextWrapWord

	cloudStatusLbl := widget.NewLabel(i18n.T(i18n.KeyCloudSyncCloudUnknown))
	localStatusBox := container.NewHBox(widget.NewLabel(cloudSyncLocalStatus(a.settings.CloudSyncLastSyncAt)))

	setLocalStatus := func(objs ...fyne.CanvasObject) {
		localStatusBox.Objects = objs
		localStatusBox.Refresh()
	}

	refreshLocalStatus := func() {
		setLocalStatus(widget.NewLabel(cloudSyncLocalStatus(a.settings.CloudSyncLastSyncAt)))
	}

	showUploadError := func(err error) {
		logPath, logErr := cloudsync.LogUploadError(err)
		if logErr != nil {
			setLocalStatus(widget.NewLabel(i18n.Tf(i18n.KeyCloudSyncUploadLogFail, logErr.Error())))
			return
		}
		link := widget.NewButton(logPath, func() { _ = fileopen.OpenPath(logPath) })
		link.Importance = widget.LowImportance
		setLocalStatus(
			widget.NewLabel(i18n.T(i18n.KeyCloudSyncUploadFailPrefix)),
			link,
		)
	}

	saveLocalConfig := func() {
		a.settings.CloudSyncAPISecret = apiSecret.Text
		a.settings.CloudSyncPassword = encPass.Text
		_ = config.SaveSettings(a.settings)
	}

	saveKeysBtn := widget.NewButton(i18n.T(i18n.KeyCloudSyncSaveLocal), func() {
		saveLocalConfig()
		dialogShowOn(parent(), i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncSaveLocalOK))
	})

	queryCloudBtn := widget.NewButton(i18n.T(i18n.KeyCloudSyncQueryCloud), func() {
		saveLocalConfig()
		if apiSecret.Text == "" {
			dialogShowOn(parent(), i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncNeedSecret))
			return
		}
		cloudStatusLbl.SetText(i18n.T(i18n.KeyCloudSyncQuerying))
		go func() {
			st, err := cloudsync.NewClient(apiSecret.Text).QueryStatus()
			fyne.Do(func() {
				if err != nil {
					cloudStatusLbl.SetText(i18n.T(i18n.KeyCloudSyncCloudUnknown))
					dialogShowErrorOn(parent(), err)
					return
				}
				if !st.Exists {
					cloudStatusLbl.SetText(i18n.T(i18n.KeyCloudSyncCloudEmpty))
					return
				}
				cloudStatusLbl.SetText(i18n.Tf(i18n.KeyCloudSyncCloudSavedAt, st.UpdatedAt))
			})
		}()
	})

	syncUpBtn := newAccentButton(i18n.T(i18n.KeyCloudSyncUpload), func() {
		saveLocalConfig()
		if apiSecret.Text == "" {
			dialogShowOn(parent(), i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncNeedSecret))
			return
		}
		if encPass.Text == "" {
			dialogShowOn(parent(), i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncNeedPassword))
			return
		}
		if len(a.store.Servers) == 0 {
			dialogShowOn(parent(), i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncNoLocalData))
			return
		}
		setLocalStatus(widget.NewLabel(i18n.T(i18n.KeyCloudSyncUploading)))
		go func() {
			updatedAt, err := cloudsync.Upload(a.store, apiSecret.Text, encPass.Text)
			fyne.Do(func() {
				if err != nil {
					showUploadError(err)
					return
				}
				now := time.Now().Format("2006-01-02 15:04:05")
				if updatedAt != "" {
					a.settings.CloudSyncLastSyncAt = updatedAt
				} else {
					a.settings.CloudSyncLastSyncAt = now
				}
				_ = config.SaveSettings(a.settings)
				refreshLocalStatus()
				if updatedAt != "" {
					cloudStatusLbl.SetText(i18n.Tf(i18n.KeyCloudSyncCloudSavedAt, updatedAt))
				}
				dialogShowOn(parent(), i18n.T(i18n.KeyCloudSyncTitle), i18n.Tf(i18n.KeyCloudSyncUploadOK, a.settings.CloudSyncLastSyncAt))
			})
		}()
	})

	syncDownBtn := newAccentButton(i18n.T(i18n.KeyCloudSyncDownload), func() {
		saveLocalConfig()
		if apiSecret.Text == "" {
			dialogShowOn(parent(), i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncNeedSecret))
			return
		}
		if encPass.Text == "" {
			dialogShowOn(parent(), i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncNeedPassword))
			return
		}
		go func() {
			plain, st, err := cloudsync.Download(apiSecret.Text, encPass.Text)
			fyne.Do(func() {
				if err != nil {
					dialogShowErrorOn(parent(), err)
					return
				}
				if !st.Exists || len(plain) == 0 {
					dialogShowOn(parent(), i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncCloudNoData))
					return
				}
				dialog.ShowConfirm(
					i18n.T(i18n.KeyCloudSyncTitle),
					i18n.T(i18n.KeyCloudSyncDownloadConfirm),
					func(ok bool) {
						if !ok {
							return
						}
						go func() {
							newStore := &config.Store{}
							if err := cloudsync.ApplyPayload(plain, newStore); err != nil {
								fyne.Do(func() { dialogShowErrorOn(parent(), err) })
								return
							}
							fyne.Do(func() {
								a.store.Servers = newStore.Servers
								a.saveServers()
								now := time.Now().Format("2006-01-02 15:04:05")
								a.settings.CloudSyncLastSyncAt = now
								_ = config.SaveSettings(a.settings)
								refreshLocalStatus()
								a.tabBar.Refresh()
								dialogShowOn(parent(), i18n.T(i18n.KeyCloudSyncTitle), i18n.Tf(i18n.KeyCloudSyncDownloadOK, now))
							})
						}()
					},
					parent(),
				)
			})
		}()
	})

	deleteBtn := newAccentButton(i18n.T(i18n.KeyCloudSyncDeleteCloud), func() {
		saveLocalConfig()
		if apiSecret.Text == "" {
			dialogShowOn(parent(), i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncNeedSecret))
			return
		}
		dialog.ShowConfirm(
			i18n.T(i18n.KeyCloudSyncDeleteTitle),
			i18n.T(i18n.KeyCloudSyncDeleteConfirm),
			func(ok bool) {
				if !ok {
					return
				}
				go func() {
					err := cloudsync.NewClient(apiSecret.Text).Delete()
					fyne.Do(func() {
						resultTitle := i18n.T(i18n.KeyCloudSyncDeleteResultTitle)
						if err != nil {
							dialogShowOn(parent(), resultTitle, i18n.Tf(i18n.KeyCloudSyncDeleteFail, err.Error()))
							return
						}
						cloudStatusLbl.SetText(i18n.T(i18n.KeyCloudSyncCloudEmpty))
						dialogShowOn(parent(), resultTitle, i18n.T(i18n.KeyCloudSyncDeleteOK))
					})
				}()
			},
			parent(),
		)
	})
	deleteBtn.Importance = widget.DangerImportance

	configSection := container.NewVBox(
		widget.NewLabel(i18n.T(i18n.KeyCloudSyncConfigSection)),
		widget.NewForm(
			widget.NewFormItem(i18n.T(i18n.KeyCloudSyncAPISecret), container.NewVBox(apiSecret, apiSecretLink)),
			widget.NewFormItem(i18n.T(i18n.KeyCloudSyncPassword), encPass),
		),
		privacyNote,
		container.NewHBox(saveKeysBtn, queryCloudBtn),
	)

	statusSection := container.NewVBox(
		widget.NewLabel(i18n.T(i18n.KeyCloudSyncStatusSection)),
		container.NewHBox(widget.NewLabel(i18n.T(i18n.KeyCloudSyncLocalStatus)), localStatusBox),
		container.NewHBox(widget.NewLabel(i18n.T(i18n.KeyCloudSyncCloudStatus)), cloudStatusLbl),
	)

	actionSection := container.NewVBox(
		widget.NewLabel(i18n.T(i18n.KeyCloudSyncActionSection)),
		container.NewHBox(syncUpBtn, syncDownBtn, layout.NewSpacer(), deleteBtn),
	)

	body := container.NewVBox(
		configSection,
		widget.NewSeparator(),
		statusSection,
		widget.NewSeparator(),
		actionSection,
	)
	cloudDlg = newModalDialogAuto(a, i18n.T(i18n.KeyCloudSyncTitle), 560, body)
	cloudDlg.SetOnClose(saveLocalConfig)
}

func cloudSyncLocalStatus(lastSync string) string {
	if lastSync == "" {
		return i18n.T(i18n.KeyCloudSyncLocalNever)
	}
	return i18n.Tf(i18n.KeyCloudSyncLocalSyncedAt, lastSync)
}
