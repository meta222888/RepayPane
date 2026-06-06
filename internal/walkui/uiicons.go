package walkui

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lxn/walk"
)

// Segoe MDL2 Assets (built into Windows 10+).
const (
	glyphRefresh = "\uE72C"
	glyphUp      = "\uE70E"
	glyphAdd     = "\uE710"
	glyphClose   = "\uE711"
	glyphDisk    = "\uEDA2"
	glyphHeart   = "\uEB51"
)

var (
	mdl2Font     *walk.Font
	mdl2FontOnce sync.Once
	iconCacheDir string
	iconCacheOnce sync.Once
)

func mdl2IconFont() *walk.Font {
	mdl2FontOnce.Do(func() {
		mdl2Font, _ = walk.NewFont("Segoe MDL2 Assets", 11, walk.FontBold)
	})
	return mdl2Font
}

func applyMDL2Font(w walk.Widget) {
	if f := mdl2IconFont(); f != nil {
		w.SetFont(f)
	}
}

func newMDL2ToolButton(parent walk.Container, glyph, tooltip string, fn func()) (*walk.ToolButton, error) {
	tb, err := walk.NewToolButton(parent)
	if err != nil {
		return nil, err
	}
	applyMDL2Font(tb)
	tb.SetText(glyph)
	if tooltip != "" {
		tb.SetToolTipText(tooltip)
	}
	if fn != nil {
		tb.Clicked().Attach(fn)
	}
	_ = tb.SetMinMaxSize(walk.Size{Width: 28, Height: 28}, walk.Size{Width: 28, Height: 28})
	return tb, nil
}

func newMDL2Label(parent walk.Container, glyph string) (*walk.Label, error) {
	lbl, err := walk.NewLabel(parent)
	if err != nil {
		return nil, err
	}
	applyMDL2Font(lbl)
	lbl.SetText(glyph)
	_ = lbl.SetMinMaxSize(walk.Size{Width: 20, Height: 0}, walk.Size{Width: 20, Height: 0})
	return lbl, nil
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

// shellIconPath returns a local path suitable for Windows SHGetFileInfo icons.
func shellIconPath(name string, isDir bool) string {
	if isDir {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return `C:\`
		}
		return home
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

func tabCompositeMaxWidth() walk.Size {
	return walk.Size{Width: 160, Height: 0}
}

func setTabCompositeActive(c *walk.Composite, active bool) {
	if c == nil {
		return
	}
	color := walk.RGB(240, 240, 240)
	if active {
		color = walk.RGB(220, 235, 252)
	}
	if brush, err := walk.NewSolidColorBrush(color); err == nil {
		c.SetBackground(brush)
	}
}

func measureTabTextWidth(text string) int {
	_ = text
	return 120
}
