package walkui

import (
	"time"

	"github.com/lxn/walk"
)

type dirEntry struct {
	name     string
	fullPath string
	size     int64
	modTime  time.Time
	isDir    bool
}

type dirModel struct {
	walk.TableModelBase
	items []dirEntry
}

func newDirModel() *dirModel {
	return &dirModel{}
}

func (m *dirModel) RowCount() int {
	return len(m.items)
}

func (m *dirModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(m.items) {
		return nil
	}
	e := m.items[row]
	switch col {
	case 0:
		return e.name
	case 1:
		if e.isDir {
			return "<DIR>"
		}
		return formatSize(e.size)
	case 2:
		return formatTime(e.modTime)
	default:
		return nil
	}
}

func (m *dirModel) setItems(items []dirEntry) {
	m.items = items
	m.PublishRowsReset()
}

func (m *dirModel) entry(row int) (dirEntry, bool) {
	if row < 0 || row >= len(m.items) {
		return dirEntry{}, false
	}
	return m.items[row], true
}
