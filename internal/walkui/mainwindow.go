package walkui

import (
	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func Run() error {
	store, err := config.Load()
	if err != nil {
		return err
	}
	settings, err := config.LoadSettings()
	if err != nil {
		settings = &config.Settings{}
	}
	initLanguage(settings)

	app := newApp(store, settings)
	app.initTransfers()
	var mw *walk.MainWindow

	if err := (MainWindow{
		AssignTo: &mw,
		Title:    i18n.T(i18n.KeyAppTitle) + " (Win32)",
		MinSize:  Size{960, 600},
		Size:     Size{1280, 760},
		Layout:   VBox{MarginsZero: true},
		OnDropFiles: app.handleDropFiles,
		MenuItems: buildMainMenus(app, mw),
		Children: []Widget{
			Composite{
				AssignTo: &app.tabBar,
				Layout:   HBox{Margins: Margins{4, 4, 4, 0}},
			},
			ToolBar{
				Items: []MenuItem{
					Action{Text: i18n.T(i18n.KeyConnect), OnTriggered: app.showConnectDialog},
					Action{Text: i18n.T(i18n.KeyRefresh), OnTriggered: func() {
						app.refreshLocal()
						app.refreshRemote()
					}},
					Action{Text: i18n.T(i18n.KeyUpload), OnTriggered: app.uploadSelected},
					Action{Text: i18n.T(i18n.KeyDownload), OnTriggered: app.downloadSelected},
				},
			},
			HSplitter{
				Children: []Widget{
					paneComposite(app, true),
					paneComposite(app, false),
				},
			},
			Composite{
				Layout: VBox{Margins: Margins{4, 4, 4, 4}},
				Children: []Widget{
					ProgressBar{AssignTo: &app.progressBar, Visible: false},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{AssignTo: &app.statusLabel, Text: i18n.T(i18n.KeyNotConnected)},
							HSpacer{},
							Label{AssignTo: &app.transferLabel, Text: i18n.T(i18n.KeyTransferIdle)},
							PushButton{
								AssignTo:  &app.reconnectBtn,
								Text:      i18n.T(i18n.KeyReconnect),
								Visible:   false,
								OnClicked: app.reconnectActiveTab,
							},
						},
					},
				},
			},
		},
	}).Create(); err != nil {
		return err
	}

	app.mw = mw
	app.attachPaneDrag(app.localTV, true)
	app.attachPaneDrag(app.remoteTV, false)
	app.refreshLocal()
	app.refreshTabBar()
	app.updateStatusBar()

	if len(store.Servers) > 0 {
		app.showConnectDialog()
	}

	mw.Run()
	return nil
}

func buildMainMenus(app *App, mw *walk.MainWindow) []MenuItem {
	return []MenuItem{
		Menu{
			Text: i18n.T(i18n.KeyMenuSettings),
			Items: []MenuItem{
				Menu{
					Text: i18n.T(i18n.KeyMenuLanguage),
					Items: []MenuItem{
						Action{Text: i18n.T(i18n.KeyMenuLangEN), OnTriggered: func() { app.setLanguage(i18n.EN) }},
						Action{Text: i18n.T(i18n.KeyMenuLangZH), OnTriggered: func() { app.setLanguage(i18n.ZH) }},
					},
				},
				Action{Text: i18n.T(i18n.KeyMenuMyServers), OnTriggered: app.showMyServers},
				Action{Text: i18n.T(i18n.KeyMenuCloudSync), OnTriggered: app.showCloudSync},
				Separator{},
				Action{Text: i18n.T(i18n.KeyMenuExit), OnTriggered: func() { mw.Close() }},
			},
		},
		Menu{
			Text: i18n.T(i18n.KeyMenuFeatures),
			Items: []MenuItem{
				Action{Text: i18n.T(i18n.KeyFeatSysInfo), OnTriggered: app.showSystemInfo},
				Action{Text: i18n.T(i18n.KeyFeatNetwork), OnTriggered: app.showNetworkInfo},
				Action{Text: i18n.T(i18n.KeyFeatDisk), OnTriggered: app.showDiskSpace},
				Action{Text: i18n.T(i18n.KeyFeatDu), OnTriggered: app.showDiskUsageTree},
				Action{Text: i18n.T(i18n.KeyFeatResources), OnTriggered: app.showResourceUsage},
				Separator{},
				Menu{
					Text: i18n.T(i18n.KeyFeatSync),
					Items: []MenuItem{
						Action{Text: i18n.T(i18n.KeyFeatSyncUp), OnTriggered: app.syncLocalToRemote},
						Action{Text: i18n.T(i18n.KeyFeatSyncDown), OnTriggered: app.syncRemoteToLocal},
					},
				},
				Separator{},
				Action{Text: i18n.T(i18n.KeyFeatShell), OnTriggered: app.showRemoteShell},
			},
		},
		Menu{
			Text: i18n.T(i18n.KeyMenuAbout),
			Items: []MenuItem{
				Action{Text: i18n.T(i18n.KeyMenuCheckUpdate), OnTriggered: app.showCheckUpdate},
				Action{Text: i18n.T(i18n.KeyMenuAboutUs), OnTriggered: app.showAboutUs},
			},
		},
	}
}

func tableColumns() []TableViewColumn {
	return []TableViewColumn{
		{Title: "Name", Width: 220},
		{Title: "Size", Width: 80},
		{Title: "Modified", Width: 140},
	}
}

func paneComposite(app *App, local bool) Widget {
	title := i18n.T(i18n.KeyRemote)
	if local {
		title = i18n.T(i18n.KeyLocal)
	}

	var pathEdit **walk.LineEdit
	var tv **walk.TableView
	var model *dirModel
	var upFn, refreshFn, activatedFn func()
	var onPathReturn func()
	var driveCombo **walk.ComboBox
	var placesCombo **walk.ComboBox

	if local {
		pathEdit = &app.localPathEdit
		tv = &app.localTV
		model = app.localModel
		upFn = app.localUp
		refreshFn = app.refreshLocal
		activatedFn = app.onLocalActivated
		driveCombo = &app.localDriveCombo
		onPathReturn = func() { app.navigateLocal(app.localPathEdit.Text()) }
	} else {
		pathEdit = &app.remotePathEdit
		tv = &app.remoteTV
		model = app.remoteModel
		upFn = app.remoteUp
		refreshFn = app.refreshRemote
		activatedFn = app.onRemoteActivated
		onPathReturn = func() { app.navigateRemote(app.remotePathEdit.Text()) }
	}

	navRow := []Widget{
		PushButton{Text: i18n.T(i18n.KeyUp), OnClicked: upFn, MaxSize: Size{48, 0}},
		PushButton{Text: i18n.T(i18n.KeyRefresh), OnClicked: refreshFn, MaxSize: Size{56, 0}},
	}
	if local {
		drives := listWindowsDrives()
		navRow = append(navRow, ComboBox{
			AssignTo: driveCombo,
			Model:    drives,
			MaxSize:  Size{56, 0},
			OnCurrentIndexChanged: func() {
				if app.localDriveCombo == nil {
					return
				}
				idx := app.localDriveCombo.CurrentIndex()
				if idx >= 0 && idx < len(drives) {
					app.navigateLocal(drives[idx])
				}
			},
		})
		places := commonPlaces()
		placeLabels := make([]string, len(places))
		for i, p := range places {
			placeLabels[i] = p.label
		}
		navRow = append(navRow, ComboBox{
			AssignTo: placesCombo,
			Model:    placeLabels,
			MaxSize:  Size{100, 0},
			OnCurrentIndexChanged: func() {
				if placesCombo == nil || *placesCombo == nil {
					return
				}
				idx := (*placesCombo).CurrentIndex()
				if idx >= 0 && idx < len(places) {
					app.navigateLocal(places[idx].path)
				}
			},
		})
	}
	navRow = append(navRow, LineEdit{
		AssignTo: pathEdit,
		OnKeyDown: func(key walk.Key) {
			if key == walk.KeyReturn {
				onPathReturn()
			}
		},
	})

	prepCtx := app.prepareLocalContextMenu
	if !local {
		prepCtx = app.prepareRemoteContextMenu
	}

	return Composite{
		Layout: VBox{Margins: Margins{4, 4, 4, 4}},
		Children: []Widget{
			Label{Text: title, Font: Font{Bold: true}},
			Composite{Layout: HBox{MarginsZero: true}, Children: navRow},
			TableView{
				AssignTo:           tv,
				AlternatingRowBG:   true,
				MultiSelection:     true,
				Columns:            tableColumns(),
				Model:              model,
				OnItemActivated:    activatedFn,
				OnMouseDown: func(x, y int, button walk.MouseButton) {
					if button == walk.RightButton {
						prepCtx()
					}
				},
			},
		},
	}
}
