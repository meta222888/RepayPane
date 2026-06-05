package walkui

type paneKind int

const (
	paneLocal paneKind = iota
	paneRemote
)

type clipItem struct {
	path  string
	name  string
	isDir bool
}

type paneClipboard struct {
	kind  paneKind
	items []clipItem
}
