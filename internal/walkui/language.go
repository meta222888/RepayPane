package walkui

import (
	"github.com/relaypane/relaypane/internal/i18n"
)

func (a *App) applyLanguage() {
	if a.mw == nil {
		return
	}
	a.mw.SetTitle(i18n.T(i18n.KeyAppTitle))
	if a.localPaneTitle != nil {
		a.localPaneTitle.SetText(i18n.T(i18n.KeyLocal))
	}
	if a.remotePaneTitle != nil {
		a.remotePaneTitle.SetText(i18n.T(i18n.KeyRemote))
	}
	if a.reconnectBtn != nil {
		a.reconnectBtn.SetText(i18n.T(i18n.KeyReconnect))
	}
	if a.toolbarConnect != nil {
		a.toolbarConnect.SetText(i18n.T(i18n.KeyConnect))
	}
	if a.toolbarRefresh != nil {
		a.toolbarRefresh.SetText(i18n.T(i18n.KeyRefresh))
	}
	if a.toolbarUpload != nil {
		a.toolbarUpload.SetText(i18n.T(i18n.KeyUpload))
	}
	if a.toolbarDownload != nil {
		a.toolbarDownload.SetText(i18n.T(i18n.KeyDownload))
	}
	if a.remoteLoadingLabel != nil {
		a.remoteLoadingLabel.SetText(i18n.T(i18n.KeyFeatLoading))
	}
	if a.remoteEmptyLabel != nil {
		a.remoteEmptyLabel.SetText(i18n.T(i18n.KeyNotConnected))
	}
	a.updateWindowTitle()
	a.refreshTabBar()
	a.updateStatusBar()
	a.refreshMainMenuLabels()
}

func (a *App) refreshMainMenuLabels() {
	menu := a.mw.Menu()
	if menu == nil || menu.Actions().Len() < 3 {
		return
	}
	settings := menu.Actions().At(0)
	features := menu.Actions().At(1)
	about := menu.Actions().At(2)
	settings.SetText(i18n.T(i18n.KeyMenuSettings))
	features.SetText(i18n.T(i18n.KeyMenuFeatures))
	about.SetText(i18n.T(i18n.KeyMenuAbout))

	if sm := settings.Menu(); sm != nil && sm.Actions().Len() >= 4 {
		sm.Actions().At(0).SetText(i18n.T(i18n.KeyMenuLanguage))
		sm.Actions().At(1).SetText(i18n.T(i18n.KeyMenuMyServers))
		sm.Actions().At(2).SetText(i18n.T(i18n.KeyMenuCloudSync))
		sm.Actions().At(3).SetText(i18n.T(i18n.KeyMenuExit))
		if lm := sm.Actions().At(0).Menu(); lm != nil && lm.Actions().Len() >= 2 {
			lm.Actions().At(0).SetText(i18n.T(i18n.KeyMenuLangEN))
			lm.Actions().At(1).SetText(i18n.T(i18n.KeyMenuLangZH))
		}
	}

	if fm := features.Menu(); fm != nil && fm.Actions().Len() >= 8 {
		fm.Actions().At(0).SetText(i18n.T(i18n.KeyFeatSysInfo))
		fm.Actions().At(1).SetText(i18n.T(i18n.KeyFeatNetwork))
		fm.Actions().At(2).SetText(i18n.T(i18n.KeyFeatDisk))
		fm.Actions().At(3).SetText(i18n.T(i18n.KeyFeatDu))
		fm.Actions().At(4).SetText(i18n.T(i18n.KeyFeatResources))
		fm.Actions().At(6).SetText(i18n.T(i18n.KeyFeatSync))
		fm.Actions().At(7).SetText(i18n.T(i18n.KeyFeatShell))
		if syncM := fm.Actions().At(6).Menu(); syncM != nil && syncM.Actions().Len() >= 2 {
			syncM.Actions().At(0).SetText(i18n.T(i18n.KeyFeatSyncUp))
			syncM.Actions().At(1).SetText(i18n.T(i18n.KeyFeatSyncDown))
		}
	}

	if am := about.Menu(); am != nil && am.Actions().Len() >= 2 {
		am.Actions().At(0).SetText(i18n.T(i18n.KeyMenuCheckUpdate))
		am.Actions().At(1).SetText(i18n.T(i18n.KeyMenuAboutUs))
	}
}
