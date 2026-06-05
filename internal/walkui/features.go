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

const netCombinedCmd = `echo "=== Interfaces (/proc/net/dev) ==="
awk 'NR>2 {
  iface=$1; gsub(":", "", iface)
  if (iface=="lo") next
  print iface "\t" $2 "\t" $10
}' /proc/net/dev 2>/dev/null
echo
echo "=== Routes ==="
(ip -o route show 2>/dev/null || ip route 2>/dev/null || route -n 2>/dev/null) | head -12
echo
echo "=== Listening Ports ==="
(ss -tulnp 2>/dev/null || netstat -tulnp 2>/dev/null || netstat -tuln 2>/dev/null)
`

const resourcesCmd = `free -b 2>/dev/null | awk '/^Mem:/ {print "MEM_TOTAL="$2; print "MEM_USED="$3; print "MEM_FREE="$4}'
grep '^cpu ' /proc/stat | awk '{idle=$5+$6; total=0; for(i=2;i<=NF;i++) total+=$i; print "CPU_IDLE=" idle; print "CPU_TOTAL=" total}'
uptime
ps -eo pid,user,comm,%cpu,%mem --sort=-%mem 2>/dev/null | head -10
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

func (a *App) showNetworkInfo() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	a.showFeatureDialog(i18n.T(i18n.KeyFeatNetwork), func(set func(string)) {
		set(i18n.T(i18n.KeyFeatLoading))
		go func() {
			out, err := client.RunCombined(netCombinedCmd)
			if err != nil && strings.TrimSpace(out) == "" {
				set(err.Error())
				return
			}
			set(strings.TrimSpace(out))
		}()
	})
}

func (a *App) showDiskSpace() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	a.showFeatureDialog(i18n.T(i18n.KeyFeatDisk), func(set func(string)) {
		set(i18n.T(i18n.KeyFeatLoading))
		go func() {
			out, err := client.RunCombined(`df -hP 2>/dev/null || df -h 2>/dev/null`)
			if err != nil && strings.TrimSpace(out) == "" {
				set(err.Error())
				return
			}
			set(strings.TrimSpace(out))
		}()
	})
}

func (a *App) showResourceUsage() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	a.showFeatureDialog(i18n.T(i18n.KeyFeatResources), func(set func(string)) {
		set(i18n.T(i18n.KeyFeatLoading))
		go func() {
			out, err := client.RunCombined(resourcesCmd)
			if err != nil && strings.TrimSpace(out) == "" {
				set(err.Error())
				return
			}
			set(strings.TrimSpace(out))
		}()
	})
}

func (a *App) showDiskUsageTree() {
	a.showDuDialog("/")
}

func (a *App) showDuDialog(dir string) {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	dir = normalizeDuPath(dir)
	a.showFeatureDialog(i18n.T(i18n.KeyFeatDu)+" — "+dir, func(set func(string)) {
		set(i18n.T(i18n.KeyFeatLoading))
		go func() {
			out, err := client.RunCombined(duListCmd(dir))
			text := formatDuOutput(out, err)
			set(text + "\n\n(double-click entry in future; use path navigation via refresh)")
		}()
	})
}

func formatDuOutput(out string, err error) string {
	if err != nil && strings.TrimSpace(out) == "" {
		return err.Error()
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var b strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			icon := "📄"
			if parts[0] == "D" {
				icon = "📁"
			}
			b.WriteString(icon + " " + parts[2] + "  " + parts[1] + " KB\n")
			continue
		}
		b.WriteString(line + "\n")
	}
	if b.Len() == 0 {
		return i18n.T(i18n.KeyFeatNoData)
	}
	return b.String()
}

func duListCmd(dir string) string {
	quoted := `"` + strings.ReplaceAll(dir, `"`, `\"`) + `"`
	tab := "\t"
	return `du -sk ` + quoted + `/* 2>/dev/null | sort -rn | while read sz p; do
  [ -z "$p" ] && continue
  if [ -d "$p" ]; then t=D; else t=F; fi
  printf "%s` + tab + `%s` + tab + `%s\n" "$t" "$sz" "$p"
done`
}

func normalizeDuPath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	cleaned := filepath.ToSlash(filepath.Clean(p))
	if cleaned == "." {
		return "/"
	}
	return cleaned
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
