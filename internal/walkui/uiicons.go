package walkui

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/relaypane/relaypane/internal/assets"

	"github.com/lxn/walk"
)

var (
	uiBmpOnce sync.Once
	bmpClose  *walk.Bitmap
	bmpNew    *walk.Bitmap
	bmpUp     *walk.Bitmap
	bmpRefresh *walk.Bitmap
	bmpDisk   *walk.Bitmap
	bmpLike   *walk.Bitmap
)

func InitUIBitmaps(dpi int) {
	uiBmpOnce.Do(func() {
		bmpClose, _ = assets.CloseBitmap(dpi)
		bmpNew, _ = assets.NewTabBitmap(dpi)
		bmpUp, _ = assets.UpBitmap(dpi)
		bmpRefresh, _ = assets.RefreshBitmap(dpi)
		bmpDisk, _ = assets.DiskBitmap(dpi)
		bmpLike, _ = assets.LikeBitmap(dpi)
	})
}

func UIBmpClose() *walk.Bitmap   { return bmpClose }
func UIBmpNew() *walk.Bitmap     { return bmpNew }
func UIBmpUp() *walk.Bitmap      { return bmpUp }
func UIBmpRefresh() *walk.Bitmap { return bmpRefresh }
func UIBmpDisk() *walk.Bitmap    { return bmpDisk }
func UIBmpLike() *walk.Bitmap    { return bmpLike }

func newPNGToolButton(parent walk.Container, bmp *walk.Bitmap, tooltip string, fn func()) (*walk.ToolButton, error) {
	return newPNGToolButtonSize(parent, bmp, tooltip, fn, tabBarHeight-2)
}

func newPNGToolButtonSize(parent walk.Container, bmp *walk.Bitmap, tooltip string, fn func(), size int) (*walk.ToolButton, error) {
	tb, err := walk.NewToolButton(parent)
	if err != nil {
		return nil, err
	}
	if bmp != nil {
		_ = tb.SetImage(bmp)
	}
	if tooltip != "" {
		tb.SetToolTipText(tooltip)
	}
	if fn != nil {
		tb.Clicked().Attach(fn)
	}
	if size <= 0 {
		size = tabBarHeight - 2
	}
	_ = tb.SetMinMaxSize(walk.Size{Width: size, Height: size}, walk.Size{Width: size, Height: size})
	return tb, nil
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}

func clearContainerChildren(c walk.Container) {
	if c == nil {
		return
	}
	for c.Children().Len() > 0 {
		w := c.Children().At(0)
		_ = c.Children().Remove(w)
		w.Dispose()
	}
}

func iconCacheRoot() string {
	iconCacheOnce.Do(func() {
		iconCacheDir = filepath.Join(os.TempDir(), "relaypane-icon-cache")
		_ = os.MkdirAll(iconCacheDir, 0o755)
	})
	return iconCacheDir
}

var (
	iconCacheDir   string
	iconCacheOnce  sync.Once
	folderIconDir  string
	folderIconOnce sync.Once
)

func genericFolderShellPath() string {
	folderIconOnce.Do(func() {
		folderIconDir = filepath.Join(iconCacheRoot(), "__folder__")
		_ = os.MkdirAll(folderIconDir, 0o755)
	})
	return folderIconDir
}

// shellIconPath returns a local path suitable for Windows SHGetFileInfo icons.
func shellIconPath(name string, isDir bool) string {
	if isDir {
		return genericFolderShellPath()
	}
	safe := strings.Map(func(r rune) rune {
		if r < 32 || r == '"' || r == '*' || r == '?' || r == '<' || r == '>' || r == '|' {
			return -1
		}
		return r
	}, name)
	if safe == "" {
		safe = "file"
	}
	if len([]rune(safe)) > 80 {
		safe = string([]rune(safe)[:80])
	}
	p := filepath.Join(iconCacheRoot(), safe)
	if _, err := os.Stat(p); err != nil {
		_ = os.WriteFile(p, nil, 0o644)
	}
	return p
}

func iconPathForEntry(e dirEntry, local bool) string {
	if local {
		return e.fullPath
	}
	return shellIconPath(e.name, e.isDir)
}

func setTabCompositeActive(c *walk.Composite, active bool) {
	if c == nil {
		return
	}
	color := colorTabInactive
	if active {
		color = colorTabActive
	}
	if brush, err := walk.NewSolidColorBrush(color); err == nil {
		c.SetBackground(brush)
	}
}

func measureTabTextWidth(text string) int {
	const btnPad = 20
	maxWidth := tabMaxWidth
	w := btnPad
	for _, r := range text {
		if r > 0xFF {
			w += 14
		} else {
			w += 8
		}
	}
	if w > maxWidth {
		return maxWidth
	}
	return w
}
