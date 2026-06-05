package ui

import (
	"sort"

	"fyne.io/fyne/v2/widget"
)

func (p *FilePane) isRowSelected(row int) bool {
	_, ok := p.selectedRows[row]
	return ok
}

func (p *FilePane) selectedRowsSorted() []int {
	rows := make([]int, 0, len(p.selectedRows))
	for row := range p.selectedRows {
		rows = append(rows, row)
	}
	sort.Ints(rows)
	return rows
}

func (p *FilePane) selectedFileRows() []int {
	rows := p.selectedRowsSorted()
	out := make([]int, 0, len(rows))
	for _, row := range rows {
		if !p.isParentRow(row) {
			out = append(out, row)
		}
	}
	return out
}

func (p *FilePane) hasFileSelection() bool {
	return len(p.selectedFileRows()) > 0
}

func (p *FilePane) selectionAnchorRow() int {
	if p.selectionAnchor >= 0 && p.isRowSelected(p.selectionAnchor) {
		return p.selectionAnchor
	}
	rows := p.selectedRowsSorted()
	if len(rows) == 0 {
		return -1
	}
	return rows[0]
}

func (p *FilePane) replaceSelection(rows map[int]struct{}) {
	if rows == nil {
		rows = make(map[int]struct{})
	}
	changed := make(map[int]struct{})
	for r := range p.selectedRows {
		changed[r] = struct{}{}
	}
	for r := range rows {
		changed[r] = struct{}{}
	}
	p.selectedRows = rows
	for r := range changed {
		if r >= 0 && r < p.rowCount() {
			p.list.RefreshItem(widget.ListItemID(r))
		}
	}
	if anchor := p.selectionAnchorRow(); anchor >= 0 {
		p.list.Select(widget.ListItemID(anchor))
	}
}

func (p *FilePane) selectRow(row int) {
	if row < 0 || row >= p.rowCount() {
		return
	}
	p.selectionAnchor = row
	p.replaceSelection(map[int]struct{}{row: {}})
}

func (p *FilePane) toggleRowSelection(row int) {
	if row < 0 || row >= p.rowCount() || p.isParentRow(row) {
		return
	}
	next := make(map[int]struct{}, len(p.selectedRows))
	for r := range p.selectedRows {
		next[r] = struct{}{}
	}
	if _, ok := next[row]; ok {
		delete(next, row)
	} else {
		next[row] = struct{}{}
	}
	p.selectionAnchor = row
	p.replaceSelection(next)
}

func (p *FilePane) selectAllFiles() {
	next := make(map[int]struct{})
	for i := 0; i < p.rowCount(); i++ {
		if !p.isParentRow(i) {
			next[i] = struct{}{}
		}
	}
	p.replaceSelection(next)
}

func (p *FilePane) noteActive() {
	p.app.activeFilePane = p
}

func (p *FilePane) rowsForClipboardAction(row int) []int {
	if p.isRowSelected(row) {
		if rows := p.selectedFileRows(); len(rows) > 0 {
			return rows
		}
	}
	if row < 0 || p.isParentRow(row) {
		return nil
	}
	return []int{row}
}

func (p *FilePane) setClipboardFromRows(rows []int) bool {
	items := make([]PaneClipItem, 0, len(rows))
	for _, row := range rows {
		if p.isParentRow(row) {
			continue
		}
		path := p.fullPathForRow(row)
		name := p.nameForRow(row)
		if path == "" || name == "" {
			continue
		}
		items = append(items, PaneClipItem{
			Path:  path,
			Name:  name,
			IsDir: p.isDirForRow(row),
		})
	}
	if len(items) == 0 {
		return false
	}
	p.app.clipboard = &PaneClipboard{Kind: p.kind, Items: items}
	return true
}
