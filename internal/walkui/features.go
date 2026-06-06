package walkui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"
	"github.com/relaypane/relaypane/internal/remote"
)

const sysInfoCmd = `echo "=== Host ==="
hostname -f 2>/dev/null || hostname
echo
echo "=== OS ==="
(uname -srmo 2>/dev/null; cat /etc/os-release 2>/dev/null | grep -E '^(PRETTY_NAME|NAME|VERSION)=' ) | head -5
echo
echo "=== Kernel ==="
uname -a
echo
echo "=== Uptime ==="
uptime
echo
echo "=== CPU ==="
(lscpu 2>/dev/null | grep -E 'Model name|Architecture|CPU\\(s\\)|Thread|Core' || grep -m1 model /proc/cpuinfo 2>/dev/null || echo "n/a")
echo
echo "=== Memory ==="
free -h 2>/dev/null || cat /proc/meminfo | head -3
echo
echo "=== Disk (summary) ==="
df -hP --output=target,size,used,avail,pcent 2>/dev/null | head -10
echo
echo "=== User ==="
whoami; id 2>/dev/null
`

func (a *App) showSystemInfo() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	a.showFeatureDialog(i18n.T(i18n.KeyFeatSysInfo), func(set func(string)) {
		set(i18n.T(i18n.KeyFeatLoading))
		go func() {
			out, err := client.RunCombined(sysInfoCmd)
			if err != nil && strings.TrimSpace(out) == "" {
				set(err.Error())
				return
			}
			set(strings.TrimSpace(out))
		}()
	})
}

func (a *App) syncLocalToRemote() {
	a.confirmSync(true)
}

func (a *App) syncRemoteToLocal() {
	a.confirmSync(false)
}

func (a *App) confirmSync(upload bool) {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	localPath := a.localPath
	remotePath := a.remotePath
	var msg string
	if upload {
		msg = i18n.Tf(i18n.KeyFeatSyncConfirmUp, localPath, remotePath)
	} else {
		msg = i18n.Tf(i18n.KeyFeatSyncConfirmDown, remotePath, localPath)
	}
	if !a.showConfirmSync(i18n.T(i18n.KeyFeatSync), msg) {
		return
	}
	if upload {
		a.runSyncUpload(client, localPath, remotePath)
	} else {
		a.runSyncDownload(client, remotePath, localPath)
	}
}

func (a *App) runSyncUpload(client *remote.Client, localDir, remoteDir string) {
	go func() {
		_ = filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(localDir, path)
			if err != nil {
				return nil
			}
			rp := filepath.ToSlash(filepath.Join(remoteDir, rel))
			a.transfers.EnqueueUpload(client, path, rp, func(err error) {
				if err != nil {
					a.showError(i18n.T(i18n.KeyFeatSync), err)
				}
			})
			return nil
		})
		a.syncUI(func() { a.refreshRemote() })
	}()
}

func (a *App) runSyncDownload(client *remote.Client, remoteDir, localDir string) {
	go func() {
		walkRemoteTree(client, remoteDir, func(p string, isDir bool) {
			if isDir {
				return
			}
			rel, err := filepath.Rel(remoteDir, filepath.FromSlash(p))
			if err != nil {
				return
			}
			lp := filepath.Join(localDir, rel)
			a.transfers.EnqueueDownload(client, p, lp, func(err error) {
				if err != nil {
					a.showError(i18n.T(i18n.KeyFeatSync), err)
				}
			})
		})
		a.syncUI(func() { a.refreshLocal() })
	}()
}

func walkRemoteTree(client *remote.Client, dir string, fn func(path string, isDir bool)) {
	entries, err := client.ListDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		fn(e.Path, e.IsDir)
		if e.IsDir {
			walkRemoteTree(client, e.Path, fn)
		}
	}
}
