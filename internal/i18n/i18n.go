package i18n

import "fmt"

type Lang string

const (
	EN Lang = "en"
	ZH Lang = "zh"
)

var current = ZH

func Current() Lang {
	return current
}

func SetLanguage(lang Lang) {
	if lang == ZH {
		current = ZH
	} else {
		current = EN
	}
}

func T(key string) string {
	if msg, ok := catalog[current][key]; ok {
		return msg
	}
	if msg, ok := catalog[EN][key]; ok {
		return msg
	}
	return key
}

func Tf(key string, args ...any) string {
	return fmt.Sprintf(T(key), args...)
}

var catalog = map[Lang]map[string]string{
	EN: enStrings,
	ZH: zhStrings,
}

const (
	KeyAppTitle          = "app.title"
	KeyMenuSettings      = "menu.settings"
	KeyMenuFeatures      = "menu.features"
	KeyMenuAbout         = "menu.about"
	KeyMenuLanguage      = "menu.language"
	KeyMenuLangEN        = "menu.lang.en"
	KeyMenuLangZH        = "menu.lang.zh"
	KeyMenuMyServers     = "menu.my_servers"
	KeyMenuCloudSync     = "menu.cloud_sync"
	KeyMenuExit          = "menu.exit"
	KeyMenuCheckUpdate   = "menu.check_update"
	KeyMenuAboutUs       = "menu.about_us"
	KeyMenuComingSoon    = "menu.coming_soon"
	KeyMenuFeaturesSoon  = "menu.features_soon"

	KeyServers      = "servers"
	KeyAddServer    = "add_server"
	KeyEdit         = "edit"
	KeyDelete       = "delete"
	KeyRefresh      = "refresh"
	KeyDisconnect   = "disconnect"
	KeyUpload       = "upload"
	KeyDownload     = "download"
	KeyLocal        = "local"
	KeyRemote       = "remote"
	KeyUp           = "up"

	KeyNotConnected       = "status.not_connected"
	KeyConnecting         = "status.connecting"
	KeyConnectionFailed   = "status.connection_failed"
	KeyConnected          = "status.connected"
	KeyDisconnected       = "status.disconnected"
	KeyConnectionLost     = "status.connection_lost"
	KeyHeartbeatSuffix    = "status.heartbeat_suffix"

	KeySelectServer     = "dialog.select_server"
	KeyChooseEdit       = "dialog.choose_edit"
	KeyChooseDelete     = "dialog.choose_delete"
	KeyDeleteConfirm    = "dialog.delete_confirm"
	KeyFileTooLarge     = "dialog.file_too_large"
	KeyFileTooLargeMsg  = "dialog.file_too_large_msg"
	KeyNotConnectedTitle = "dialog.not_connected"
	KeyNotConnectedUpload = "dialog.not_connected_upload"
	KeyNotConnectedFirst  = "dialog.not_connected_first"
	KeySelectFile       = "dialog.select_file"
	KeySelectLocalUpload  = "dialog.select_local_upload"
	KeySelectRemoteDownload = "dialog.select_remote_download"
	KeyInvalidLocalPath = "dialog.invalid_local_path"

	KeyServerFormTitle = "form.server_title"
	KeySave            = "form.save"
	KeyOK              = "form.ok"
	KeyCancel          = "form.cancel"
	KeyFormName        = "form.name"
	KeyFormHost        = "form.host"
	KeyFormPort        = "form.port"
	KeyFormUsername    = "form.username"
	KeyFormPassword    = "form.password"
	KeyFormPrivateKey  = "form.private_key"
	KeyFormAutoSSHKey  = "form.auto_ssh_key"
	KeyFormSelectKey   = "form.select_key"
	KeyFormKeySelected = "form.key_selected"
	KeyFormKeyNone     = "form.key_none"
	KeyFormSaveConnect = "form.save_connect"
	KeyFormConnectOnly = "form.connect_only"
	KeyFormRequired    = "form.required"
	KeyFormRemoteRoot  = "form.remote_root"
	KeyFormHeartbeat   = "form.heartbeat"

	KeySaved         = "editor.saved"
	KeySavedMsg      = "editor.saved_msg"
	KeySaveFailed    = "editor.save_failed"
	KeyEditorSaving  = "editor.saving"
	KeySaveSuccessTitle = "editor.save_success_title"
	KeySaveSuccessAt    = "editor.save_success_at"
	KeySaveFailedTitle  = "editor.save_failed_title"
	KeySaveFailedAt     = "editor.save_failed_at"
	KeyEditorRevert     = "editor.revert"
	KeyEditorRevertConfirm = "editor.revert_confirm"
	KeyEditorRevertFailed  = "editor.revert_failed"
	KeyUnsaved       = "editor.unsaved"
	KeyDiscard       = "editor.discard"
	KeyEditTitle     = "editor.title"
	KeyCtrlSSave     = "editor.ctrl_s"
	KeyNotConnectedErr = "editor.not_connected"

	KeyCheckUpdateTitle = "about.check_update_title"
	KeyCheckUpdateMsg   = "about.check_update_msg"
	KeyCheckUpdateChecking = "about.check_update_checking"
	KeyCheckUpdateLatest   = "about.check_update_latest"
	KeyCheckUpdateAvailable = "about.check_update_available"
	KeyCheckUpdateDownload  = "about.check_update_download"
	KeyCheckUpdateFailed    = "about.check_update_failed"
	KeyAboutTitle       = "about.title"
	KeyAboutIntro       = "about.intro"
	KeyAboutVersion     = "about.version"
	KeyAboutWebsite     = "about.website"
	KeyMyServersTitle   = "about.my_servers_title"
	KeyMyServersHint    = "about.my_servers_hint"

	KeyColName       = "col.name"
	KeyColSize       = "col.size"
	KeyColModified   = "col.modified"
	KeyColPermissions = "col.permissions"
	KeyPanelLocal     = "pane.panel_local"
	KeyPanelRemote    = "pane.panel_remote"
	KeyLocalHeader   = "pane.local_header"
	KeyRemoteHeader  = "pane.remote_header"
	KeyStatusBarConnected = "status.bar_connected"
	KeyStatusConnected    = "status.connected_to"
	KeyStatusSyncing      = "status.syncing"
	KeyStatusQueue        = "status.queue_label"
	KeyTransferIdle  = "status.transfer_idle"
	KeyQueue         = "status.queue"
	KeyNewFolder     = "action.new_folder"
	KeyCloseTab      = "action.close_tab"
	KeyNewTab        = "action.new_tab"
	KeyNewTabConnect = "action.new_tab_connect"
	KeyReconnect     = "action.reconnect"

	KeySidebarPlaces   = "sidebar.places"
	KeySidebarDrive    = "sidebar.drive"
	KeyPlaceHome       = "place.home"
	KeyPlaceDesktop    = "place.desktop"
	KeyPlaceDocuments  = "place.documents"
	KeyPlacePictures   = "place.pictures"
	KeyPlaceDownloads  = "place.downloads"
	KeyPlaceMusic      = "place.music"
	KeyPlaceVideos     = "place.videos"

	KeyConnectPickerTitle = "connect.picker_title"
	KeyConnectPickerHint  = "connect.picker_hint"
	KeyConnect            = "connect.action"
	KeyNewConnection      = "connect.new"
	KeyPassphraseTitle    = "passphrase.title"
	KeyPassphrasePrompt   = "passphrase.prompt"
	KeyPassphraseHint     = "passphrase.hint"
	KeyKeyPickerTitle     = "keypicker.title"
	KeyKeyPickerHint      = "keypicker.hint"

	KeyCtxCopy            = "ctx.copy"
	KeyCtxPaste           = "ctx.paste"
	KeyCtxDelete          = "ctx.delete"
	KeyCtxNewFolder       = "ctx.new_folder"
	KeyCtxNewFile         = "ctx.new_file"
	KeyDeleteFileConfirm  = "dialog.delete_file_confirm"
	KeyFileExists         = "dialog.file_exists"
	KeyFileExistsTitle    = "dialog.file_exists_title"
	KeyFileExistsConflict = "dialog.file_exists_conflict"
	KeyOverwrite          = "dialog.overwrite"
	KeyRename             = "dialog.rename"
	KeyRenamePrompt       = "dialog.rename_prompt"
	KeyCtrlSSaveLocal     = "editor.ctrl_s_local"

	KeyFeatLoading        = "feat.loading"
	KeyFeatNoData         = "feat.no_data"
	KeyFeatSysInfo        = "feat.sysinfo"
	KeyFeatNetwork        = "feat.network"
	KeyFeatDisk           = "feat.disk"
	KeyFeatDu             = "feat.du"
	KeyFeatResources      = "feat.resources"
	KeyFeatSync           = "feat.sync"
	KeyFeatSyncUp         = "feat.sync_up"
	KeyFeatSyncDown       = "feat.sync_down"
	KeyFeatSyncConfirmUp  = "feat.sync_confirm_up"
	KeyFeatSyncConfirmDown = "feat.sync_confirm_down"
	KeyFeatShell          = "feat.shell"
	KeyFeatShellHint      = "feat.shell_hint"
	KeyFeatShellHistory   = "feat.shell_history"
	KeyFeatShellNoHistory = "feat.shell_no_history"
	KeyFeatShellRun       = "feat.shell_run"
	KeyFeatShellDelOne    = "feat.shell_del_one"
	KeyFeatShellDelAll      = "feat.shell_del_all"
	KeyFeatShellInteractive = "feat.shell_interactive"
	KeyFeatShellExitCode    = "feat.shell_exit_code"
	KeyFeatShellTimeout     = "feat.shell_timeout"
	KeyFeatShellCopyHint    = "feat.shell_copy_hint"
	KeyFeatRunning        = "feat.running"
	KeyFeatNetTraffic     = "feat.net_traffic"
	KeyFeatNetPorts       = "feat.net_ports"
	KeyFeatNetAutoRefresh = "feat.net_auto_refresh"
	KeyFeatRefreshTraffic = "feat.refresh_traffic"
	KeyFeatRefreshPorts   = "feat.refresh_ports"
	KeyFeatNetIfaceDetail = "feat.net_iface_detail"
	KeyFeatNetRate        = "feat.net_rate"
	KeyFeatNetRouting     = "feat.net_routing"
	KeyFeatNetSinceBoot   = "feat.net_since_boot"
	KeyFeatDiskDetail     = "feat.disk_detail"
	KeyFeatResCPU         = "feat.res_cpu"
	KeyFeatResMemory      = "feat.res_memory"
	KeyFeatResUptime      = "feat.res_uptime"
	KeyFeatResProcesses   = "feat.res_processes"
	KeyFeatResCPUPct      = "feat.res_cpu_pct"
	KeyFeatResMemDetail   = "feat.res_mem_detail"

	KeyCloudSyncTitle           = "cloudsync.title"
	KeyCloudSyncConfigSection   = "cloudsync.config_section"
	KeyCloudSyncStatusSection   = "cloudsync.status_section"
	KeyCloudSyncActionSection   = "cloudsync.action_section"
	KeyCloudSyncAPISecret       = "cloudsync.api_secret"
	KeyCloudSyncAPISecretHint   = "cloudsync.api_secret_hint"
	KeyCloudSyncAPISecretLink   = "cloudsync.api_secret_link"
	KeyCloudSyncPassword        = "cloudsync.password"
	KeyCloudSyncPasswordHint    = "cloudsync.password_hint"
	KeyCloudSyncPrivacyNote     = "cloudsync.privacy_note"
	KeyCloudSyncQueryCloud      = "cloudsync.query_cloud"
	KeyCloudSyncLocalStatus     = "cloudsync.local_status"
	KeyCloudSyncCloudStatus     = "cloudsync.cloud_status"
	KeyCloudSyncLocalNever      = "cloudsync.local_never"
	KeyCloudSyncLocalSyncedAt   = "cloudsync.local_synced_at"
	KeyCloudSyncCloudUnknown    = "cloudsync.cloud_unknown"
	KeyCloudSyncCloudEmpty      = "cloudsync.cloud_empty"
	KeyCloudSyncCloudSavedAt    = "cloudsync.cloud_saved_at"
	KeyCloudSyncQuerying        = "cloudsync.querying"
	KeyCloudSyncUpload          = "cloudsync.upload"
	KeyCloudSyncDownload        = "cloudsync.download"
	KeyCloudSyncDeleteCloud     = "cloudsync.delete_cloud"
	KeyCloudSyncDeleteTitle     = "cloudsync.delete_title"
	KeyCloudSyncDeleteConfirm   = "cloudsync.delete_confirm"
	KeyCloudSyncDeleteOK        = "cloudsync.delete_ok"
	KeyCloudSyncNeedSecret      = "cloudsync.need_secret"
	KeyCloudSyncNeedPassword    = "cloudsync.need_password"
	KeyCloudSyncNoLocalData     = "cloudsync.no_local_data"
	KeyCloudSyncUploading       = "cloudsync.uploading"
	KeyCloudSyncUploadOK        = "cloudsync.upload_ok"
	KeyCloudSyncDownloadConfirm = "cloudsync.download_confirm"
	KeyCloudSyncDownloadOK      = "cloudsync.download_ok"
	KeyCloudSyncCloudNoData     = "cloudsync.cloud_no_data"
)

var enStrings = map[string]string{
	KeyAppTitle:       "RelayPane",
	KeyMenuSettings:   "Settings",
	KeyMenuFeatures:   "Features",
	KeyMenuAbout:      "About",
	KeyMenuLanguage:   "Language",
	KeyMenuLangEN:     "English",
	KeyMenuLangZH:     "中文",
	KeyMenuMyServers:  "My Servers",
	KeyMenuCloudSync:  "Cloud Sync",
	KeyMenuExit:       "Exit",
	KeyMenuCheckUpdate: "Check for Updates",
	KeyMenuAboutUs:    "About Us",
	KeyMenuComingSoon: "Coming Soon",
	KeyMenuFeaturesSoon: "More features are under development.",

	KeyServers:    "Servers",
	KeyAddServer:  "Add Server",
	KeyEdit:       "Edit",
	KeyDelete:     "Delete",
	KeyRefresh:    "Refresh",
	KeyDisconnect: "Disconnect",
	KeyUpload:     "Upload  →",
	KeyDownload:   "←  Download",
	KeyLocal:      "Local",
	KeyRemote:     "Remote",
	KeyUp:         "Up",

	KeyNotConnected:     "Not connected",
	KeyConnecting:       "Connecting to %s…",
	KeyConnectionFailed: "Connection failed",
	KeyConnected:        "Connected: %s (%s@%s)",
	KeyDisconnected:     "Disconnected",
	KeyConnectionLost:   "Connection lost (heartbeat failed)",
	KeyHeartbeatSuffix:  " · heartbeat %ds",

	KeySelectServer:         "Select server",
	KeyChooseEdit:           "Choose a server to edit.",
	KeyChooseDelete:         "Choose a server to delete.",
	KeyDeleteConfirm:        "Delete %q?",
	KeyFileTooLarge:         "File too large",
	KeyFileTooLargeMsg:      "%s is %.1f MB. Open anyway? (Ctrl+S saves back to server)",
	KeyNotConnectedTitle:    "Not connected",
	KeyNotConnectedUpload:   "Connect to a server before uploading.",
	KeyNotConnectedFirst:    "Connect to a server first.",
	KeySelectFile:           "Select a file",
	KeySelectLocalUpload:    "Select a local file to upload.",
	KeySelectRemoteDownload: "Select a remote file to download.",
	KeyInvalidLocalPath:     "invalid local path: %s",

	KeyServerFormTitle: "Server",
	KeySave:            "Save",
	KeyOK:              "OK",
	KeyCancel:          "Cancel",
	KeyFormName:        "Name",
	KeyFormHost:        "Host",
	KeyFormPort:        "Port",
	KeyFormUsername:    "Username",
	KeyFormPassword:    "Password",
	KeyFormPrivateKey:  "Private key file",
	KeyFormAutoSSHKey:  "Auto (~/.ssh)",
	KeyFormSelectKey:   "Select key…",
	KeyFormKeySelected: "Selected: %s",
	KeyFormKeyNone:     "No key selected",
	KeyFormSaveConnect: "Save & Connect",
	KeyFormConnectOnly: "Connect only",
	KeyFormRequired:    "Host and username are required.",
	KeyFormRemoteRoot:  "Remote root",
	KeyFormHeartbeat:   "Heartbeat (seconds, 0=off)",

	KeySaved:           "Saved",
	KeySavedMsg:        "Saved to %s",
	KeySaveFailed:      "Save failed: %s",
	KeyEditorSaving:    "Saving…",
	KeySaveSuccessTitle: "Save successful",
	KeySaveSuccessAt:    "Save successful [%s]",
	KeySaveFailedTitle:  "Save failed",
	KeySaveFailedAt:     "Save failed: %s [%s]",
	KeyEditorRevert:     "Revert",
	KeyEditorRevertConfirm: "Discard unsaved changes and reload from file?",
	KeyEditorRevertFailed:  "Reload failed: %s",
	KeyUnsaved:         "Unsaved changes",
	KeyDiscard:         "Discard changes?",
	KeyEditTitle:       "Edit: %s",
	KeyCtrlSSave:       "Ctrl+S to save to server",
	KeyNotConnectedErr: "not connected",

	KeyCheckUpdateTitle: "Check for Updates",
	KeyCheckUpdateMsg:   "Automatic update will be available in a future release.",
	KeyCheckUpdateChecking: "Checking for updates…",
	KeyCheckUpdateLatest:   "You are on the latest version (%s).",
	KeyCheckUpdateAvailable: "A new version is available: %s (current: %s)",
	KeyCheckUpdateDownload:  "Download update",
	KeyCheckUpdateFailed:    "Update check failed: %s",
	KeyAboutTitle:       "About RelayPane",
	KeyAboutIntro:       "RelayPane is a lightweight SFTP client for transferring and editing files between your computer and remote servers.",
	KeyAboutVersion:     "Version %s",
	KeyAboutWebsite:     "https://pc530.com",
	KeyMyServersTitle:   "My Servers",
	KeyMyServersHint:    "Select a saved server (double-click to connect), or manage your server list.",

	KeyColName:        "Name",
	KeyColSize:        "Size",
	KeyColModified:    "Modified",
	KeyColPermissions: "Permissions",
	KeyPanelLocal:     "Local",
	KeyPanelRemote:    "Remote",
	KeyLocalHeader:    "Local — %s",
	KeyRemoteHeader:   "Remote — %s",
	KeyStatusBarConnected: "Connected to %s",
	KeyStatusConnected:    "Connected to",
	KeyStatusSyncing:      "Silently syncing to server…",
	KeyStatusQueue:        "Queue: %d",
	KeyTransferIdle:   "— MB/s",
	KeyQueue:          "Queue: %d",
	KeyNewFolder:      "New Folder",
	KeyCloseTab:       "Close session",
	KeyNewTab:         "New connection",
	KeyNewTabConnect:  "+ Connect",
	KeyReconnect:      "Reconnect",

	KeySidebarPlaces:  "Places",
	KeySidebarDrive:   "Drive",
	KeyPlaceHome:      "Home",
	KeyPlaceDesktop:   "Desktop",
	KeyPlaceDocuments: "Documents",
	KeyPlacePictures:  "Pictures",
	KeyPlaceDownloads: "Downloads",
	KeyPlaceMusic:     "Music",
	KeyPlaceVideos:    "Videos",

	KeyConnectPickerTitle: "Connect to Server",
	KeyConnectPickerHint:  "Select a saved server (double-click to connect), or create a new connection.",
	KeyConnect:            "Connect",
	KeyNewConnection:      "New connection…",
	KeyPassphraseTitle:    "Private Key Passphrase",
	KeyPassphrasePrompt:   "Enter passphrase",
	KeyPassphraseHint:     "Your private key file is encrypted. Enter the passphrase to continue.",
	KeyKeyPickerTitle:     "Select Private Key",
	KeyKeyPickerHint:      "Pick a key from ~/.ssh or enter the full path below.",

	KeyCtxCopy:           "Copy",
	KeyCtxPaste:          "Paste",
	KeyCtxDelete:         "Delete",
	KeyCtxNewFolder:      "New Folder",
	KeyCtxNewFile:        "New File",
	KeyDeleteFileConfirm: "Delete %q?",
	KeyFileExists:        "%q already exists",
	KeyFileExistsTitle:   "File Already Exists",
	KeyFileExistsConflict: "%q already exists.\nChoose an action:",
	KeyOverwrite:         "Overwrite",
	KeyRename:            "Rename",
	KeyRenamePrompt:      "Enter a new name:",
	KeyCtrlSSaveLocal:    "Ctrl+S to save",

	KeyFeatLoading:        "Loading…",
	KeyFeatNoData:         "No data",
	KeyFeatSysInfo:        "System Info",
	KeyFeatNetwork:        "Network Info",
	KeyFeatDisk:           "Disk Space",
	KeyFeatDu:             "Directory Usage",
	KeyFeatResources:      "CPU & Memory",
	KeyFeatSync:           "Sync",
	KeyFeatSyncUp:         "Sync local → server",
	KeyFeatSyncDown:       "Sync server → local",
	KeyFeatSyncConfirmUp:  "Sync local directory to server?\n\nLocal:\n%s\n\nRemote:\n%s",
	KeyFeatSyncConfirmDown: "Sync server directory to local?\n\nRemote:\n%s\n\nLocal:\n%s",
	KeyFeatShell:          "Remote Command (Ctrl+E)",
	KeyFeatShellHint:      "Enter shell command…",
	KeyFeatShellHistory:   "History:",
	KeyFeatShellNoHistory: "No history",
	KeyFeatShellRun:       "Run",
	KeyFeatShellDelOne:    "Delete selected",
	KeyFeatShellDelAll:      "Clear all",
	KeyFeatShellInteractive: "Interactive commands (vim, top, less, etc.) are not supported in this window. Try: cat, head, tail, sed -n, grep.",
	KeyFeatShellExitCode:    "exit code",
	KeyFeatShellTimeout:     "command timed out (90s); interactive programs may hang",
	KeyFeatShellCopyHint:    "Select output text to copy (Ctrl+C).",
	KeyFeatRunning:        "Running…",
	KeyFeatNetTraffic:     "Traffic",
	KeyFeatNetPorts:       "Open Ports",
	KeyFeatNetAutoRefresh: "Refresh every 5s",
	KeyFeatRefreshTraffic: "Refresh traffic",
	KeyFeatRefreshPorts:   "Refresh ports",
	KeyFeatNetIfaceDetail: "↓ Received %s  ·  ↑ Sent %s",
	KeyFeatNetRate:        "Current speed: ↓ %s  ·  ↑ %s",
	KeyFeatNetRouting:     "Routing",
	KeyFeatNetSinceBoot:   "Cumulative since boot",
	KeyFeatDiskDetail:     "Used %s / %s  ·  Free %s  ·  %s",
	KeyFeatResCPU:         "Processor",
	KeyFeatResMemory:      "Memory",
	KeyFeatResUptime:      "Uptime",
	KeyFeatResProcesses:   "Top processes (by memory)",
	KeyFeatResCPUPct:      "In use: %.0f%%",
	KeyFeatResMemDetail:   "Used %s / %s (%.0f%%)",

	KeyCloudSyncTitle:           "Cloud Sync",
	KeyCloudSyncConfigSection:   "Configuration",
	KeyCloudSyncStatusSection:   "Status",
	KeyCloudSyncActionSection:   "Actions",
	KeyCloudSyncAPISecret:       "EasyStorage API Secret",
	KeyCloudSyncAPISecretHint:   "From pc530.com EasyStorage console",
	KeyCloudSyncAPISecretLink:   "Get my API Secret for free: https://pc530.com/easystorage/",
	KeyCloudSyncPassword:        "Encryption password",
	KeyCloudSyncPasswordHint:    "Required; any length (even one character)",
	KeyCloudSyncPrivacyNote:     "Server configs and private key contents are encrypted before upload. EasyStorage API Secret and encryption password are stored on this device only — never uploaded to the cloud.",
	KeyCloudSyncQueryCloud:      "Query cloud status",
	KeyCloudSyncLocalStatus:     "This device:",
	KeyCloudSyncCloudStatus:     "Cloud:",
	KeyCloudSyncLocalNever:      "Not synced yet",
	KeyCloudSyncLocalSyncedAt:   "Last synced at %s",
	KeyCloudSyncCloudUnknown:    "Not queried",
	KeyCloudSyncCloudEmpty:      "No data on cloud",
	KeyCloudSyncCloudSavedAt:    "Last saved on cloud: %s",
	KeyCloudSyncQuerying:        "Querying…",
	KeyCloudSyncUpload:          "Sync to cloud",
	KeyCloudSyncDownload:        "Sync from cloud",
	KeyCloudSyncDeleteCloud:     "Delete cloud data",
	KeyCloudSyncDeleteTitle:     "Delete cloud data",
	KeyCloudSyncDeleteConfirm:   "This permanently deletes cloud server data. It cannot be recovered. Continue?",
	KeyCloudSyncDeleteOK:        "Cloud data deleted.",
	KeyCloudSyncNeedSecret:      "Please enter your EasyStorage API Secret.",
	KeyCloudSyncNeedPassword:    "Please enter an encryption password (cannot be empty).",
	KeyCloudSyncNoLocalData:     "No local server database to upload.",
	KeyCloudSyncUploading:       "Uploading to cloud…",
	KeyCloudSyncUploadOK:        "Synced to cloud. Local last sync: %s",
	KeyCloudSyncDownloadConfirm: "This will overwrite local server data with the cloud copy. Continue?",
	KeyCloudSyncDownloadOK:      "Synced from cloud at %s",
	KeyCloudSyncCloudNoData:     "No data on cloud.",
}

var zhStrings = map[string]string{
	KeyAppTitle:       "RelayPane",
	KeyMenuSettings:   "设置",
	KeyMenuFeatures:   "功能",
	KeyMenuAbout:      "关于",
	KeyMenuLanguage:   "语言",
	KeyMenuLangEN:     "English",
	KeyMenuLangZH:     "中文",
	KeyMenuMyServers:  "我的服务器",
	KeyMenuCloudSync:  "云端同步",
	KeyMenuExit:       "退出",
	KeyMenuCheckUpdate: "检查更新",
	KeyMenuAboutUs:    "关于我们",
	KeyMenuComingSoon: "即将推出",
	KeyMenuFeaturesSoon: "更多功能正在开发中。",

	KeyServers:    "服务器",
	KeyAddServer:  "添加服务器",
	KeyEdit:       "编辑",
	KeyDelete:     "删除",
	KeyRefresh:    "刷新",
	KeyDisconnect: "断开连接",
	KeyUpload:     "上传  →",
	KeyDownload:   "←  下载",
	KeyLocal:      "本地",
	KeyRemote:     "远程",
	KeyUp:         "上级",

	KeyNotConnected:     "未连接",
	KeyConnecting:       "正在连接 %s…",
	KeyConnectionFailed: "连接失败",
	KeyConnected:        "已连接：%s（%s@%s）",
	KeyDisconnected:     "已断开",
	KeyConnectionLost:   "连接已断开（心跳失败）",
	KeyHeartbeatSuffix:  " · 心跳 %d 秒",

	KeySelectServer:         "选择服务器",
	KeyChooseEdit:           "请选择要编辑的服务器。",
	KeyChooseDelete:         "请选择要删除的服务器。",
	KeyDeleteConfirm:        "确定删除 %q？",
	KeyFileTooLarge:         "文件过大",
	KeyFileTooLargeMsg:      "%s 为 %.1f MB，仍要打开吗？（Ctrl+S 可保存回服务器）",
	KeyNotConnectedTitle:    "未连接",
	KeyNotConnectedUpload:   "请先连接服务器再上传。",
	KeyNotConnectedFirst:    "请先连接服务器。",
	KeySelectFile:           "选择文件",
	KeySelectLocalUpload:    "请选择要上传的本地文件。",
	KeySelectRemoteDownload: "请选择要下载的远程文件。",
	KeyInvalidLocalPath:     "无效的本地路径：%s",

	KeyServerFormTitle: "服务器",
	KeySave:            "保存",
	KeyOK:              "确定",
	KeyCancel:          "取消",
	KeyFormName:        "名称",
	KeyFormHost:        "主机",
	KeyFormPort:        "端口",
	KeyFormUsername:    "用户名",
	KeyFormPassword:    "密码",
	KeyFormPrivateKey:  "私钥文件",
	KeyFormAutoSSHKey:  "自动（使用 ~/.ssh 目录私钥）",
	KeyFormSelectKey:   "选择私钥…",
	KeyFormKeySelected: "已选：%s",
	KeyFormKeyNone:     "未选择私钥",
	KeyFormSaveConnect: "保存并连接",
	KeyFormConnectOnly: "仅连接",
	KeyFormRequired:    "请填写主机和用户名。",
	KeyFormRemoteRoot:  "远程根目录",
	KeyFormHeartbeat:   "心跳间隔（秒，0=关闭）",

	KeySaved:           "已保存",
	KeySavedMsg:        "已保存到 %s",
	KeySaveFailed:      "保存失败：%s",
	KeyEditorSaving:    "保存中…",
	KeySaveSuccessTitle: "保存成功",
	KeySaveSuccessAt:    "保存成功 [%s]",
	KeySaveFailedTitle:  "保存失败",
	KeySaveFailedAt:     "保存失败：%s [%s]",
	KeyEditorRevert:     "还原",
	KeyEditorRevertConfirm: "放弃未保存的更改并从文件重新读取？",
	KeyEditorRevertFailed:  "读取失败：%s",
	KeyUnsaved:         "未保存的更改",
	KeyDiscard:         "放弃更改？",
	KeyEditTitle:       "编辑：%s",
	KeyCtrlSSave:       "Ctrl+S 保存到服务器",
	KeyNotConnectedErr: "未连接",

	KeyCheckUpdateTitle: "检查更新",
	KeyCheckUpdateMsg:   "自动更新功能将在后续版本中提供。",
	KeyCheckUpdateChecking: "正在检查更新…",
	KeyCheckUpdateLatest:   "当前已是最新版本（%s）。",
	KeyCheckUpdateAvailable: "发现新版本：%s（当前：%s）",
	KeyCheckUpdateDownload:  "立即下载更新",
	KeyCheckUpdateFailed:    "检查更新失败：%s",
	KeyAboutTitle:       "关于 RelayPane",
	KeyAboutIntro:       "RelayPane 是一款轻量级 SFTP 客户端，用于在本地与远程服务器之间传输和编辑文件。",
	KeyAboutVersion:     "版本 %s",
	KeyAboutWebsite:     "https://pc530.com",
	KeyMyServersTitle:   "我的服务器",
	KeyMyServersHint:    "选择已保存的服务器（双击可连接），或管理服务器列表。",

	KeyColName:        "名称",
	KeyColSize:        "大小",
	KeyColModified:    "修改日期",
	KeyColPermissions: "权限",
	KeyPanelLocal:     "本地",
	KeyPanelRemote:    "远程",
	KeyLocalHeader:    "本地 — %s",
	KeyRemoteHeader:   "远程 — %s",
	KeyStatusBarConnected: "已连接到 %s",
	KeyStatusConnected:    "已连接到",
	KeyStatusSyncing:      "正在静默同步至服务器…",
	KeyStatusQueue:        "队列: %d",
	KeyTransferIdle:   "— MB/s",
	KeyQueue:          "队列: %d",
	KeyNewFolder:      "新建文件夹",
	KeyCloseTab:       "关闭会话",
	KeyNewTab:         "新建连接",
	KeyNewTabConnect:  "+连接",
	KeyReconnect:      "重新连接",

	KeySidebarPlaces:  "常用目录",
	KeySidebarDrive:   "磁盘",
	KeyPlaceHome:      "主目录",
	KeyPlaceDesktop:   "桌面",
	KeyPlaceDocuments: "我的文档",
	KeyPlacePictures:  "我的图片",
	KeyPlaceDownloads: "下载",
	KeyPlaceMusic:     "音乐",
	KeyPlaceVideos:    "视频",

	KeyConnectPickerTitle: "连接服务器",
	KeyConnectPickerHint:  "选择已保存的服务器（双击可连接），或新建连接。",
	KeyConnect:            "连接",
	KeyNewConnection:      "新建连接…",
	KeyPassphraseTitle:    "私钥密码",
	KeyPassphrasePrompt:   "请输入私钥密码",
	KeyPassphraseHint:     "您的私钥文件已加密，请输入密码后继续连接。",
	KeyKeyPickerTitle:     "选择私钥",
	KeyKeyPickerHint:      "从 ~/.ssh 列表中选择，或在下方输入完整路径。",

	KeyCtxCopy:           "复制",
	KeyCtxPaste:          "粘贴",
	KeyCtxDelete:         "删除",
	KeyCtxNewFolder:      "新建目录",
	KeyCtxNewFile:        "新建文件",
	KeyDeleteFileConfirm: "确定删除 %q？",
	KeyFileExists:        "%q 已存在",
	KeyFileExistsTitle:   "文件已存在",
	KeyFileExistsConflict: "%q 已存在。\n请选择操作：",
	KeyOverwrite:         "覆盖",
	KeyRename:            "重命名",
	KeyRenamePrompt:      "请输入新名称：",
	KeyCtrlSSaveLocal:    "Ctrl+S 保存",

	KeyFeatLoading:        "加载中…",
	KeyFeatNoData:         "无数据",
	KeyFeatSysInfo:        "系统信息",
	KeyFeatNetwork:        "网络信息",
	KeyFeatDisk:           "磁盘空间",
	KeyFeatDu:             "详细占用空间",
	KeyFeatResources:      "CPU、内存使用",
	KeyFeatSync:           "同步",
	KeyFeatSyncUp:         "同步本地目录到服务器",
	KeyFeatSyncDown:       "同步服务器目录到本地",
	KeyFeatSyncConfirmUp:  "确定将本地目录同步到服务器？\n\n本地：\n%s\n\n远程：\n%s",
	KeyFeatSyncConfirmDown: "确定将服务器目录同步到本地？\n\n远程：\n%s\n\n本地：\n%s",
	KeyFeatShell:          "执行远程命令 (Ctrl+E)",
	KeyFeatShellHint:      "输入 shell 命令…",
	KeyFeatShellHistory:   "历史：",
	KeyFeatShellNoHistory: "无历史记录",
	KeyFeatShellRun:       "执行",
	KeyFeatShellDelOne:    "删除选中",
	KeyFeatShellDelAll:      "全部删除",
	KeyFeatShellInteractive: "不支持交互式命令（vim、top、less 等）。可改用：cat、head、tail、sed -n、grep。",
	KeyFeatShellExitCode:    "退出码",
	KeyFeatShellTimeout:     "命令超时（90 秒）；交互程序可能会卡住",
	KeyFeatShellCopyHint:    "可选中输出文本后 Ctrl+C 复制。",
	KeyFeatRunning:        "执行中…",
	KeyFeatNetTraffic:     "网络流量",
	KeyFeatNetPorts:       "端口开放",
	KeyFeatNetAutoRefresh: "每 5 秒刷新",
	KeyFeatRefreshTraffic: "刷新流量",
	KeyFeatRefreshPorts:   "刷新端口",
	KeyFeatNetIfaceDetail: "↓ 接收 %s  ·  ↑ 发送 %s",
	KeyFeatNetRate:        "当前速率：↓ %s  ·  ↑ %s",
	KeyFeatNetRouting:     "路由",
	KeyFeatNetSinceBoot:   "自启动以来累计",
	KeyFeatDiskDetail:     "已用 %s / %s  ·  可用 %s  ·  %s",
	KeyFeatResCPU:         "处理器",
	KeyFeatResMemory:      "内存",
	KeyFeatResUptime:      "运行时间",
	KeyFeatResProcesses:   "占用最高的进程（按内存）",
	KeyFeatResCPUPct:      "使用率：%.0f%%",
	KeyFeatResMemDetail:   "已用 %s / %s（%.0f%%）",

	KeyCloudSyncTitle:           "云端同步",
	KeyCloudSyncConfigSection:   "配置",
	KeyCloudSyncStatusSection:   "状态",
	KeyCloudSyncActionSection:   "操作",
	KeyCloudSyncAPISecret:       "易储 API Secret",
	KeyCloudSyncAPISecretHint:   "从 pc530.com 易储控制台获取",
	KeyCloudSyncAPISecretLink:   "免费获取我的 API Secret: https://pc530.com/easystorage/",
	KeyCloudSyncPassword:        "加密密码",
	KeyCloudSyncPasswordHint:    "必填，长度不限（一位也可以）",
	KeyCloudSyncPrivacyNote:     "上传前会对服务器配置与私钥内容加密，没有密码无法解密。易储 API Secret 与加密密码仅保存在本机，不会上传到云端。",
	KeyCloudSyncQueryCloud:      "查询云端状态",
	KeyCloudSyncLocalStatus:     "本机：",
	KeyCloudSyncCloudStatus:     "云端：",
	KeyCloudSyncLocalNever:      "未同步",
	KeyCloudSyncLocalSyncedAt:   "上次同步：%s",
	KeyCloudSyncCloudUnknown:    "未查询",
	KeyCloudSyncCloudEmpty:      "云端无数据",
	KeyCloudSyncCloudSavedAt:    "云端上次保存：%s",
	KeyCloudSyncQuerying:        "查询中…",
	KeyCloudSyncUpload:          "同步到云端",
	KeyCloudSyncDownload:        "从云端同步",
	KeyCloudSyncDeleteCloud:     "删除云端数据",
	KeyCloudSyncDeleteTitle:     "删除云端数据",
	KeyCloudSyncDeleteConfirm:   "将永久删除云端服务器数据，不可恢复。确定继续？",
	KeyCloudSyncDeleteOK:        "云端数据已删除。",
	KeyCloudSyncNeedSecret:      "请填写易储 API Secret。",
	KeyCloudSyncNeedPassword:    "请填写加密密码（不能为空）。",
	KeyCloudSyncNoLocalData:     "本地没有服务器数据库。",
	KeyCloudSyncUploading:       "正在上传到云端…",
	KeyCloudSyncUploadOK:        "已同步到云端。本机上次同步：%s",
	KeyCloudSyncDownloadConfirm: "将从云端覆盖本地服务器数据，确定继续？",
	KeyCloudSyncDownloadOK:      "已从云端同步 [%s]",
	KeyCloudSyncCloudNoData:     "云端无数据。",
}
