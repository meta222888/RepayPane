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
	KeyFormRemoteRoot  = "form.remote_root"
	KeyFormHeartbeat   = "form.heartbeat"

	KeySaved         = "editor.saved"
	KeySavedMsg      = "editor.saved_msg"
	KeySaveFailed    = "editor.save_failed"
	KeyUnsaved       = "editor.unsaved"
	KeyDiscard       = "editor.discard"
	KeyEditTitle     = "editor.title"
	KeyCtrlSSave     = "editor.ctrl_s"
	KeyNotConnectedErr = "editor.not_connected"

	KeyCheckUpdateTitle = "about.check_update_title"
	KeyCheckUpdateMsg   = "about.check_update_msg"
	KeyAboutTitle       = "about.title"
	KeyAboutIntro       = "about.intro"
	KeyAboutVersion     = "about.version"
	KeyAboutWebsite     = "about.website"
	KeyMyServersTitle   = "about.my_servers_title"

	KeyColName       = "col.name"
	KeyColSize       = "col.size"
	KeyColModified   = "col.modified"
	KeyColPermissions = "col.permissions"
	KeyLocalHeader   = "pane.local_header"
	KeyRemoteHeader  = "pane.remote_header"
	KeyStatusBarConnected = "status.bar_connected"
	KeyTransferIdle  = "status.transfer_idle"
	KeyQueue         = "status.queue"
	KeyNewFolder     = "action.new_folder"
	KeyCloseTab      = "action.close_tab"
	KeyNewTab        = "action.new_tab"

	KeySidebarPlaces   = "sidebar.places"
	KeySidebarDrive    = "sidebar.drive"
	KeyPlaceHome       = "place.home"
	KeyPlaceDesktop    = "place.desktop"
	KeyPlaceDocuments  = "place.documents"
	KeyPlacePictures   = "place.pictures"
	KeyPlaceDownloads  = "place.downloads"
	KeyPlaceMusic      = "place.music"
	KeyPlaceVideos     = "place.videos"
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
	KeyFormRemoteRoot:  "Remote root",
	KeyFormHeartbeat:   "Heartbeat (seconds, 0=off)",

	KeySaved:           "Saved",
	KeySavedMsg:        "Saved to %s",
	KeySaveFailed:      "Save failed: %s",
	KeyUnsaved:         "Unsaved changes",
	KeyDiscard:         "Discard changes?",
	KeyEditTitle:       "Edit: %s",
	KeyCtrlSSave:       "Ctrl+S to save to server",
	KeyNotConnectedErr: "not connected",

	KeyCheckUpdateTitle: "Check for Updates",
	KeyCheckUpdateMsg:   "Automatic update will be available in a future release.",
	KeyAboutTitle:       "About RelayPane",
	KeyAboutIntro:       "RelayPane is a lightweight SFTP client for transferring and editing files between your computer and remote servers.",
	KeyAboutVersion:     "Version %s",
	KeyAboutWebsite:     "https://pc530.com",
	KeyMyServersTitle:   "My Servers",

	KeyColName:        "Name",
	KeyColSize:        "Size",
	KeyColModified:    "Modified",
	KeyColPermissions: "Permissions",
	KeyLocalHeader:    "Local — %s",
	KeyRemoteHeader:   "Remote — %s",
	KeyStatusBarConnected: "Connected to %s",
	KeyTransferIdle:   "— MB/s",
	KeyQueue:          "Queue: %d",
	KeyNewFolder:      "New Folder",
	KeyCloseTab:       "Close session",
	KeyNewTab:         "New connection",

	KeySidebarPlaces:  "Places",
	KeySidebarDrive:   "Drive",
	KeyPlaceHome:      "Home",
	KeyPlaceDesktop:   "Desktop",
	KeyPlaceDocuments: "Documents",
	KeyPlacePictures:  "Pictures",
	KeyPlaceDownloads: "Downloads",
	KeyPlaceMusic:     "Music",
	KeyPlaceVideos:    "Videos",
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
	KeyFormRemoteRoot:  "远程根目录",
	KeyFormHeartbeat:   "心跳间隔（秒，0=关闭）",

	KeySaved:           "已保存",
	KeySavedMsg:        "已保存到 %s",
	KeySaveFailed:      "保存失败：%s",
	KeyUnsaved:         "未保存的更改",
	KeyDiscard:         "放弃更改？",
	KeyEditTitle:       "编辑：%s",
	KeyCtrlSSave:       "Ctrl+S 保存到服务器",
	KeyNotConnectedErr: "未连接",

	KeyCheckUpdateTitle: "检查更新",
	KeyCheckUpdateMsg:   "自动更新功能将在后续版本中提供。",
	KeyAboutTitle:       "关于 RelayPane",
	KeyAboutIntro:       "RelayPane 是一款轻量级 SFTP 客户端，用于在本地与远程服务器之间传输和编辑文件。",
	KeyAboutVersion:     "版本 %s",
	KeyAboutWebsite:     "https://pc530.com",
	KeyMyServersTitle:   "我的服务器",

	KeyColName:        "名称",
	KeyColSize:        "大小",
	KeyColModified:    "修改日期",
	KeyColPermissions: "权限",
	KeyLocalHeader:    "本地 — %s",
	KeyRemoteHeader:   "远程 — %s",
	KeyStatusBarConnected: "已连接到 %s",
	KeyTransferIdle:   "— MB/s",
	KeyQueue:          "队列: %d",
	KeyNewFolder:      "新建文件夹",
	KeyCloseTab:       "关闭会话",
	KeyNewTab:         "新建连接",

	KeySidebarPlaces:  "常用目录",
	KeySidebarDrive:   "磁盘",
	KeyPlaceHome:      "主目录",
	KeyPlaceDesktop:   "桌面",
	KeyPlaceDocuments: "我的文档",
	KeyPlacePictures:  "我的图片",
	KeyPlaceDownloads: "下载",
	KeyPlaceMusic:     "音乐",
	KeyPlaceVideos:    "视频",
}
