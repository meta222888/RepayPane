package kernbridge

import (
	"os"
	"strings"

	"github.com/relaypane/relaypane/internal/config"
	"github.com/relaypane/relaypane/internal/i18n"
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

func (a *App) ShowSystemInfo() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	a.host.showFeatureDialog(i18n.T(i18n.KeyFeatSysInfo), func(set func(string)) {
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

func (a *App) ShowDiskSpace() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	a.host.showFeatureDialog(i18n.T(i18n.KeyFeatDisk), func(set func(string)) {
		set(i18n.T(i18n.KeyFeatLoading))
		go func() {
			out, err := client.RunCombined("df -hP 2>/dev/null || df -h")
			if err != nil && strings.TrimSpace(out) == "" {
				set(err.Error())
				return
			}
			set(strings.TrimSpace(out))
		}()
	})
}

func (a *App) ShowResourceUsage() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	cmd := `free -b 2>/dev/null | awk '/^Mem:/{print "MEM_TOTAL="$2,"MEM_USED="$3}'
head -1 /proc/stat
uptime
ps -eo pid,user,comm,%cpu,%mem --sort=-%mem 2>/dev/null | head -10`
	a.host.showFeatureDialog(i18n.T(i18n.KeyFeatResources), func(set func(string)) {
		set(i18n.T(i18n.KeyFeatLoading))
		go func() {
			out, err := client.RunCombined(cmd)
			if err != nil && strings.TrimSpace(out) == "" {
				set(err.Error())
				return
			}
			set(strings.TrimSpace(out))
		}()
	})
}

func (a *App) ShowNetworkInfo() {
	client, ok := a.requireClient()
	if !ok {
		return
	}
	cmd := `echo "=== /proc/net/dev ==="
cat /proc/net/dev 2>/dev/null | head -20
echo
echo "=== Routes ==="
(ip route 2>/dev/null || route -n 2>/dev/null) | head -12
echo
echo "=== Listening ==="
(ss -tulnp 2>/dev/null || netstat -tulnp 2>/dev/null || netstat -tuln 2>/dev/null) | head -15`
	a.host.showFeatureDialog(i18n.T(i18n.KeyFeatNetwork), func(set func(string)) {
		set(i18n.T(i18n.KeyFeatLoading))
		go func() {
			out, err := client.RunCombined(cmd)
			if err != nil && strings.TrimSpace(out) == "" {
				set(err.Error())
				return
			}
			set(strings.TrimSpace(out))
		}()
	})
}

func (a *App) ShowDuTree(dir string) string {
	client, ok := a.requireClient()
	if !ok {
		return mustJSON(map[string]string{"error": "not connected"})
	}
	if dir == "" {
		dir = "/"
	}
	cmd := "du -sk '" + shellQuote(dir) + "'/* 2>/dev/null | sort -rn | head -50"
	out, err := client.RunCombined(cmd)
	if err != nil && strings.TrimSpace(out) == "" {
		return mustJSON(map[string]string{"error": err.Error()})
	}
	return mustJSON(map[string]string{"dir": dir, "output": strings.TrimSpace(out)})
}

func shellQuote(s string) string {
	return strings.ReplaceAll(s, "'", "'\\''")
}

func (a *App) ShowRemoteShell() {
	a.host.showFeatureDialog(i18n.T(i18n.KeyFeatShell), func(set func(string)) {
		set(i18n.T(i18n.KeyFeatShellHint))
	})
}

func (a *App) AboutJSON() string {
	return mustJSON(map[string]string{
		"title":   i18n.T(i18n.KeyAboutTitle),
		"message": i18n.T(i18n.KeyAboutIntro),
	})
}

func (a *App) I18nJSON(key string) string {
	return i18n.T(key)
}

func listLocalDirEnsure(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (a *App) SetLocalDrive(drive string) {
	if len(drive) >= 2 && drive[1] == ':' {
		a.NavigateLocal(drive + `\`)
	}
}

func (a *App) GetShellHistoryJSON() string {
	return mustJSON(a.settings.ShellHistory)
}

func (a *App) ClearShellHistory() {
	a.settings.ShellHistory = nil
	_ = config.SaveSettings(a.settings)
}
