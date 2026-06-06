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
	initFileIcons()
	InitUIBitmaps(96)

	app := newApp(store, settings)
	app.initTransfers()
	var mw *walk.MainWindow

	if err := (MainWindow{
		AssignTo: &mw,
		Title:    i18n.T(i18n.KeyAppTitle),
		MinSize:  Size{winMinW, winMinH},
		Size:     Size{winDefaultW, winDefaultH},
		Layout:   VBox{MarginsZero: true},
		Font:     uiFont(),
		OnDropFiles: app.handleDropFiles,
		MenuItems: buildMainMenus(app, mw),
		Children: []Widget{
			Composite{
				AssignTo: &app.tabBar,
				MinSize:  fixedHeight(tabBarHeight),
				MaxSize:  fixedHeight(tabBarHeight),
				Layout:   HBox{Margins: Margins{Left: 4, Top: 2, Right: 4, Bottom: 0}, Spacing: 0},
			},
			ToolBar{
				MinSize: fixedHeight(toolBarHeight),
				MaxSize: fixedHeight(toolBarHeight),
				Items: []MenuItem{
					Action{AssignTo: &app.toolbarConnect, Text: i18n.T(i18n.KeyConnect), OnTriggered: app.showConnectDialog},
					Action{AssignTo: &app.toolbarRefresh, Text: i18n.T(i18n.KeyRefresh), OnTriggered: func() {
						app.refreshLocal()
						app.refreshRemote()
					}},
					Separator{},
					Action{AssignTo: &app.toolbarUpload, Text: i18n.T(i18n.KeyUpload), OnTriggered: app.uploadSelected},
					Action{AssignTo: &app.toolbarDownload, Text: i18n.T(i18n.KeyDownload), OnTriggered: app.downloadSelected},
				},
			},
			HSplitter{
				Children: []Widget{
					paneComposite(app, true),
					paneComposite(app, false),
				},
			},
			Composite{
				MinSize: fixedHeight(statusBarH),
				MaxSize: fixedHeight(statusBarH),
				Layout:  HBox{Margins: Margins{Left: 6, Top: 4, Right: 6, Bottom: 4}, Spacing: 10},
				Children: []Widget{
					Label{
						AssignTo:  &app.connDot,
						Text:      "●",
						TextColor: colorDisconnected,
						Font:      Font{Family: "Segoe UI", PointSize: 8, Bold: true},
						MinSize:   Size{statusDotBar, statusDotBar},
						MaxSize:   Size{statusDotBar, statusDotBar},
					},
					Label{AssignTo: &app.statusLabel, Text: i18n.T(i18n.KeyNotConnected)},
					Label{
						AssignTo:  &app.transferFileLabel,
						Text:      "",
						TextColor: colorTextMuted,
						Visible:   false,
					},
					ProgressBar{
						AssignTo: &app.progressBar,
						MinSize:  Size{progressBarW, progressBarH},
						MaxSize:  Size{progressBarW, progressBarH},
						Visible:  false,
					},
					Label{
						AssignTo:  &app.transferPctLabel,
						Text:      "",
						TextColor: colorTextMuted,
						Visible:   false,
					},
					HSpacer{},
					Label{
						AssignTo:  &app.transferLabel,
						Text:      i18n.T(i18n.KeyTransferIdle),
						Font:      monoFont(),
					},
					PushButton{
						AssignTo:  &app.reconnectBtn,
						Text:      i18n.T(i18n.KeyReconnect),
						Visible:   false,
						OnClicked: app.reconnectActiveTab,
					},
				},
			},
		},
	}).Create(); err != nil {
		return err
	}

	app.mw = mw
	app.applyWindowIcon(mw)
	app.setupRenameEdit(app.localRenameEdit, true)
	app.setupRenameEdit(app.remoteRenameEdit, false)
	app.attachPaneDrag(app.localTV, true)
	app.attachPaneDrag(app.remoteTV, false)
	app.registerShortcuts()
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

func paneComposite(app *App, local bool) Widget {
	var titleLabel **walk.Label
	var pathEdit **walk.LineEdit
	var tv **walk.TableView
	var model *dirModel
	var renamePanel **walk.Composite
	var renameEdit **walk.LineEdit
	var loadingLabel **walk.Label
	var emptyLabel **walk.Label
	var upFn, refreshFn, activatedFn func()
	var onPathReturn func()
	var driveCombo **walk.ComboBox
	var placesCombo **walk.ComboBox

	if local {
		titleLabel = &app.localPaneTitle
		pathEdit = &app.localPathEdit
		tv = &app.localTV
		model = app.localModel
		renamePanel = &app.localRenamePanel
		renameEdit = &app.localRenameEdit
		upFn = app.localUp
		refreshFn = app.refreshLocal
		activatedFn = app.onLocalActivated
		driveCombo = &app.localDriveCombo
		onPathReturn = func() { app.navigateLocal(app.localPathEdit.Text()) }
	} else {
		titleLabel = &app.remotePaneTitle
		pathEdit = &app.remotePathEdit
		tv = &app.remoteTV
		model = app.remoteModel
		renamePanel = &app.remoteRenamePanel
		renameEdit = &app.remoteRenameEdit
		loadingLabel = &app.remoteLoadingLabel
		emptyLabel = &app.remoteEmptyLabel
		upFn = app.remoteUp
		refreshFn = app.refreshRemote
		activatedFn = app.onRemoteActivated
		onPathReturn = func() { app.navigateRemote(app.remotePathEdit.Text()) }
	}

	navRow := []Widget{
		navIconButton(UIBmpUp(), i18n.T(i18n.KeyUp), upFn),
		navIconButton(UIBmpRefresh(), i18n.T(i18n.KeyRefresh), refreshFn),
	}
	if local {
		drives := listWindowsDrives()
		navRow = append(navRow,
			markIconView(UIBmpDisk()),
			ComboBox{
				AssignTo: driveCombo,
				Model:    drives,
				MinSize:  Size{driveComboW, 0},
				MaxSize:  Size{driveComboW, 0},
				OnCurrentIndexChanged: func() {
					if app.localDriveCombo == nil {
						return
					}
					idx := app.localDriveCombo.CurrentIndex()
					if idx >= 0 && idx < len(drives) {
						app.navigateLocal(drives[idx])
					}
				},
			},
		)
		places := commonPlaces()
		placeLabels := make([]string, len(places))
		for i, p := range places {
			placeLabels[i] = p.label
		}
		navRow = append(navRow,
			markIconView(UIBmpLike()),
			ComboBox{
				AssignTo: placesCombo,
				Model:    placeLabels,
				MinSize:  Size{placesComboW, 0},
				MaxSize:  Size{placesComboW, 0},
				OnCurrentIndexChanged: func() {
					if placesCombo == nil || *placesCombo == nil {
						return
					}
					idx := (*placesCombo).CurrentIndex()
					if idx >= 0 && idx < len(places) {
						app.navigateLocal(places[idx].path)
					}
				},
			},
		)
	} else {
		navRow = append(navRow, remoteNavSpacer())
	}
	navRow = append(navRow, LineEdit{
		AssignTo: pathEdit,
		MinSize:  Size{0, 20},
		OnKeyDown: func(key walk.Key) {
			if key == walk.KeyReturn {
				onPathReturn()
			}
		},
	})

	prepCtx := app.prepareLocalContextMenu
	paneTitle := i18n.T(i18n.KeyLocal)
	if !local {
		prepCtx = app.prepareRemoteContextMenu
		paneTitle = i18n.T(i18n.KeyRemote)
	}

	children := []Widget{
		Label{
			AssignTo: titleLabel,
			Text:     paneTitle,
			Font:     uiFontBold(),
			MinSize:  fixedHeight(paneTitleH),
			MaxSize:  fixedHeight(paneTitleH),
		},
		Composite{
			MinSize:  fixedHeight(navRowHeight),
			MaxSize:  fixedHeight(navRowHeight),
			Layout:   HBox{Margins: Margins{Left: 2, Top: 0, Right: 2, Bottom: 0}, Spacing: 4},
			Children: navRow,
		},
		Composite{
			AssignTo: renamePanel,
			Layout:   HBox{},
			MinSize:  fixedHeight(navRowHeight),
			MaxSize:  fixedHeight(navRowHeight),
			Visible:  false,
			Children: []Widget{
				Label{Text: i18n.T(i18n.KeyRename) + ":"},
				LineEdit{AssignTo: renameEdit},
			},
		},
	}
	if !local {
		children = append(children,
			Label{
				AssignTo:  emptyLabel,
				Text:      i18n.T(i18n.KeyNotConnected),
				Font:      Font{Family: "Segoe UI", PointSize: 9, Italic: true},
				TextColor: colorRemoteEmpty,
				Visible:   true,
			},
			Label{AssignTo: loadingLabel, Text: i18n.T(i18n.KeyFeatLoading), Visible: false},
		)
	}
	children = append(children, TableView{
		AssignTo:         tv,
		AlternatingRowBG: true,
		MultiSelection:   true,
		Columns:          tableColumns(),
		Model:            model,
		StyleCell:        paneStyleCell(local, model),
		OnItemActivated:  activatedFn,
		OnMouseDown: func(x, y int, button walk.MouseButton) {
			if button == walk.RightButton {
				prepCtx()
			}
		},
	})

	return Composite{
		Layout:   VBox{Margins: paneMargins()},
		Children: children,
	}
}
