package walkui

import (
	"strings"

	"github.com/relaypane/relaypane/internal/i18n"

	"github.com/lxn/walk"
)

type netIfaceCardUI struct {
	root   *walk.Composite
	detail *walk.Label
}

type netTrafficPanel struct {
	box      *walk.Composite
	cards    map[string]*netIfaceCardUI
	routeTE  *walk.TextEdit
	routeSet bool
}

func newNetTrafficPanel(box *walk.Composite) *netTrafficPanel {
	return &netTrafficPanel{box: box, cards: make(map[string]*netIfaceCardUI)}
}

func (p *netTrafficPanel) card(name string) *netIfaceCardUI {
	if c, ok := p.cards[name]; ok {
		return c
	}
	root, err := walk.NewComposite(p.box)
	if err != nil {
		return nil
	}
	vl := walk.NewVBoxLayout()
	vl.SetMargins(walk.Margins{0, 0, 0, 6})
	_ = root.SetLayout(vl)

	title, err := walk.NewLabel(root)
	if err != nil {
		root.Dispose()
		return nil
	}
	font, _ := walk.NewFont("Segoe UI", 9, walk.FontBold)
	if font != nil {
		title.SetFont(font)
	}
	title.SetText(name)

	detail, err := walk.NewLabel(root)
	if err != nil {
		root.Dispose()
		return nil
	}

	c := &netIfaceCardUI{root: root, detail: detail}
	p.cards[name] = c
	return c
}

func (p *netTrafficPanel) setRoute(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		if p.routeTE != nil {
			p.routeTE.SetVisible(false)
		}
		return
	}
	if p.routeTE == nil {
		title, _ := walk.NewLabel(p.box)
		if title != nil {
			font, _ := walk.NewFont("Segoe UI", 9, walk.FontBold)
			if font != nil {
				title.SetFont(font)
			}
			title.SetText(i18n.T(i18n.KeyFeatNetRouting))
		}
		te, err := walk.NewTextEdit(p.box)
		if err != nil {
			return
		}
		_ = te.SetReadOnly(true)
		font, _ := walk.NewFont("Consolas", 9, 0)
		if font != nil {
			te.SetFont(font)
		}
		p.routeTE = te
	}
	p.routeTE.SetVisible(true)
	setMultilineText(p.routeTE, text)
	p.routeSet = true
}

func (p *netTrafficPanel) update(stats []netIfaceStat, routeOut string, firstBuild bool) {
	if p.box == nil {
		return
	}
	if firstBuild && len(p.cards) == 0 && len(stats) == 0 {
		clearComposite(p.box)
		addLabel(p.box, i18n.T(i18n.KeyFeatNoData))
		return
	}
	if firstBuild && len(p.cards) == 0 {
		clearComposite(p.box)
	}
	seen := make(map[string]bool, len(stats))
	for _, s := range stats {
		c := p.card(s.name)
		if c == nil {
			continue
		}
		seen[s.name] = true
		c.root.SetVisible(true)
		c.detail.SetText(i18n.Tf(i18n.KeyFeatNetIfaceDetail, formatBytes(s.rx), formatBytes(s.tx)))
	}
	for name, c := range p.cards {
		if !seen[name] {
			c.root.SetVisible(false)
		}
	}
	if firstBuild {
		p.setRoute(routeOut)
	}
}

func (p *netTrafficPanel) updateInPlace(stats []netIfaceStat) {
	if p.box == nil {
		return
	}
	seen := make(map[string]bool, len(stats))
	for _, s := range stats {
		c := p.card(s.name)
		if c == nil {
			continue
		}
		seen[s.name] = true
		c.root.SetVisible(true)
		c.detail.SetText(i18n.Tf(i18n.KeyFeatNetIfaceDetail, formatBytes(s.rx), formatBytes(s.tx)))
	}
	for name, c := range p.cards {
		if !seen[name] {
			c.root.SetVisible(false)
		}
	}
}
