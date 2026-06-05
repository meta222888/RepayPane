package walkui

import (
	"os/exec"
	"time"

	"github.com/relaypane/relaypane/internal/cloudsync"
	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/update"
	"github.com/relaypane/relaypane/internal/version"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (a *App) showAboutUs() {
	msg := i18n.T(i18n.KeyAboutIntro) + "\n\n" +
		i18n.Tf(i18n.KeyAboutVersion, version.Version) + "\n" +
		i18n.T(i18n.KeyAboutWebsite)
	a.showMsg(i18n.T(i18n.KeyAboutTitle), msg)
}

func (a *App) showCheckUpdate() {
	var dlg *walk.Dialog
	var status *walk.Label

	if err := (Dialog{
		AssignTo: &dlg,
		Title:    i18n.T(i18n.KeyCheckUpdateTitle),
		MinSize:  Size{460, 180},
		Layout:   VBox{},
		Children: []Widget{
			Label{AssignTo: &status, Text: i18n.T(i18n.KeyCheckUpdateChecking)},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{Text: i18n.T(i18n.KeyOK), OnClicked: func() { dlg.Cancel() }},
				},
			},
		},
	}).Create(a.mw); err != nil {
		return
	}

	go func() {
		rel, err := update.FetchLatestRelease()
		a.syncUI(func() {
			if err != nil {
				status.SetText(i18n.Tf(i18n.KeyCheckUpdateFailed, err.Error()))
				return
			}
			current := version.Version
			if update.IsNewer(current, rel.Version) {
				status.SetText(i18n.Tf(i18n.KeyCheckUpdateAvailable, rel.Version, current))
				_ = openURL(rel.HTMLURL)
				return
			}
			status.SetText(i18n.Tf(i18n.KeyCheckUpdateLatest, current))
		})
	}()

	dlg.Run()
}

func openURL(u string) error {
	return exec.Command("cmd", "/C", "start", "", u).Start()
}

func (a *App) showCloudSync() {
	var dlg *walk.Dialog
	var apiEdit, passEdit *walk.LineEdit
	var cloudLbl, localLbl *walk.Label

	saveLocalConfig := func() {
		a.settings.CloudSyncAPISecret = apiEdit.Text()
		a.settings.CloudSyncPassword = passEdit.Text()
		_ = config.SaveSettings(a.settings)
	}

	refreshLocalStatus := func() {
		if localLbl != nil {
			localLbl.SetText(cloudSyncLocalStatus(a.settings.CloudSyncLastSyncAt))
		}
	}

	_, _ = Dialog{
		AssignTo: &dlg,
		Title:    i18n.T(i18n.KeyCloudSyncTitle),
		MinSize:  Size{560, 480},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: i18n.T(i18n.KeyCloudSyncConfigSection), Font: Font{Bold: true}},
			Label{Text: i18n.T(i18n.KeyCloudSyncAPISecret)},
			LineEdit{AssignTo: &apiEdit, Text: a.settings.CloudSyncAPISecret, PasswordMode: true},
			Label{Text: i18n.T(i18n.KeyCloudSyncPassword)},
			LineEdit{AssignTo: &passEdit, Text: a.settings.CloudSyncPassword, PasswordMode: true},
			Label{Text: i18n.T(i18n.KeyCloudSyncPrivacyNote)},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{Text: i18n.T(i18n.KeyCloudSyncSaveLocal), OnClicked: func() {
						saveLocalConfig()
						a.showMsg(i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncSaveLocalOK))
					}},
					PushButton{Text: i18n.T(i18n.KeyCloudSyncQueryCloud), OnClicked: func() {
						saveLocalConfig()
						if apiEdit.Text() == "" {
							a.showMsg(i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncNeedSecret))
							return
						}
						cloudLbl.SetText(i18n.T(i18n.KeyCloudSyncQuerying))
						go func() {
							st, err := cloudsync.NewClient(apiEdit.Text()).QueryStatus()
							a.syncUI(func() {
								if err != nil {
									cloudLbl.SetText(i18n.T(i18n.KeyCloudSyncCloudUnknown))
									a.showError(i18n.T(i18n.KeyCloudSyncTitle), err)
									return
								}
								if !st.Exists {
									cloudLbl.SetText(i18n.T(i18n.KeyCloudSyncCloudEmpty))
									return
								}
								cloudLbl.SetText(i18n.Tf(i18n.KeyCloudSyncCloudSavedAt, st.UpdatedAt))
							})
						}()
					}},
				},
			},
			Label{Text: i18n.T(i18n.KeyCloudSyncStatusSection), Font: Font{Bold: true}},
			Composite{Layout: HBox{}, Children: []Widget{
				Label{Text: i18n.T(i18n.KeyCloudSyncLocalStatus)},
				Label{AssignTo: &localLbl, Text: cloudSyncLocalStatus(a.settings.CloudSyncLastSyncAt)},
			}},
			Composite{Layout: HBox{}, Children: []Widget{
				Label{Text: i18n.T(i18n.KeyCloudSyncCloudStatus)},
				Label{AssignTo: &cloudLbl, Text: i18n.T(i18n.KeyCloudSyncCloudUnknown)},
			}},
			Label{Text: i18n.T(i18n.KeyCloudSyncActionSection), Font: Font{Bold: true}},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{Text: i18n.T(i18n.KeyCloudSyncUpload), OnClicked: func() {
						saveLocalConfig()
						if apiEdit.Text() == "" || passEdit.Text() == "" {
							a.showMsg(i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncNeedPassword))
							return
						}
						if len(a.store.Servers) == 0 {
							a.showMsg(i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncNoLocalData))
							return
						}
						go func() {
							updatedAt, err := cloudsync.Upload(a.store, apiEdit.Text(), passEdit.Text())
							a.syncUI(func() {
								if err != nil {
									a.showError(i18n.T(i18n.KeyCloudSyncTitle), err)
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
								a.showMsg(i18n.T(i18n.KeyCloudSyncTitle), i18n.Tf(i18n.KeyCloudSyncUploadOK, a.settings.CloudSyncLastSyncAt))
							})
						}()
					}},
					PushButton{Text: i18n.T(i18n.KeyCloudSyncDownload), OnClicked: func() {
						saveLocalConfig()
						if apiEdit.Text() == "" || passEdit.Text() == "" {
							a.showMsg(i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncNeedPassword))
							return
						}
						go func() {
							plain, st, err := cloudsync.Download(apiEdit.Text(), passEdit.Text())
							a.syncUI(func() {
								if err != nil {
									a.showError(i18n.T(i18n.KeyCloudSyncTitle), err)
									return
								}
								if !st.Exists || len(plain) == 0 {
									a.showMsg(i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncCloudNoData))
									return
								}
								if !a.showConfirmSync(i18n.T(i18n.KeyCloudSyncTitle), i18n.T(i18n.KeyCloudSyncDownloadConfirm)) {
									return
								}
								go func() {
									newStore := &config.Store{}
									if err := cloudsync.ApplyPayload(plain, newStore); err != nil {
										a.showError(i18n.T(i18n.KeyCloudSyncTitle), err)
										return
									}
									a.syncUI(func() {
										a.store.Servers = newStore.Servers
										_ = a.saveServers()
										now := time.Now().Format("2006-01-02 15:04:05")
										a.settings.CloudSyncLastSyncAt = now
										_ = config.SaveSettings(a.settings)
										refreshLocalStatus()
										a.refreshTabBar()
										a.showMsg(i18n.T(i18n.KeyCloudSyncTitle), i18n.Tf(i18n.KeyCloudSyncDownloadOK, now))
									})
								}()
							})
						}()
					}},
					PushButton{Text: i18n.T(i18n.KeyCloudSyncDeleteCloud), OnClicked: func() {
						saveLocalConfig()
						if apiEdit.Text() == "" {
							return
						}
						if !a.showConfirmSync(i18n.T(i18n.KeyCloudSyncDeleteTitle), i18n.T(i18n.KeyCloudSyncDeleteConfirm)) {
							return
						}
						go func() {
							err := cloudsync.NewClient(apiEdit.Text()).Delete()
							a.syncUI(func() {
								if err != nil {
									a.showMsg(i18n.T(i18n.KeyCloudSyncDeleteResultTitle), i18n.Tf(i18n.KeyCloudSyncDeleteFail, err.Error()))
									return
								}
								cloudLbl.SetText(i18n.T(i18n.KeyCloudSyncCloudEmpty))
								a.showMsg(i18n.T(i18n.KeyCloudSyncDeleteResultTitle), i18n.T(i18n.KeyCloudSyncDeleteOK))
							})
						}()
					}},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{Text: i18n.T(i18n.KeyOK), OnClicked: func() {
						saveLocalConfig()
						dlg.Cancel()
					}},
				},
			},
		},
	}.Run(a.mw)
}

func cloudSyncLocalStatus(lastSync string) string {
	if lastSync == "" {
		return i18n.T(i18n.KeyCloudSyncLocalNever)
	}
	return i18n.Tf(i18n.KeyCloudSyncLocalSyncedAt, lastSync)
}
