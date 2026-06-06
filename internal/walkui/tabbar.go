package walkui

import (
	"github.com/relaypane/relaypane/internal/i18n"

	"github.com/lxn/walk"
)

func (a *App) refreshTabBar() {
	if a.tabBar == nil {
		return
	}
	a.syncUI(func() {
		clearContainerChildren(a.tabBar)

		for i, tab := range a.tabs {
			idx := i
			active := i == a.activeTab
			label := tab.tabLabel()
			switch tab.state {
			case tabConnecting:
				label += " …"
			case tabDisconnected:
				if tab.client == nil {
					label += " !"
				}
			}

			wrap, err := walk.NewComposite(a.tabBar)
			if err != nil {
				continue
			}
			hl := walk.NewHBoxLayout()
			hl.SetMargins(walk.Margins{2, 2, 2, 2})
			hl.SetSpacing(2)
			if err := wrap.SetLayout(hl); err != nil {
				wrap.Dispose()
				continue
			}
			_ = wrap.SetMinMaxSize(walk.Size{Width: 0, Height: 0}, tabCompositeMaxWidth())

			tabBtn, err := walk.NewPushButton(wrap)
			if err != nil {
				wrap.Dispose()
				continue
			}
			tabBtn.SetText(truncateRunes(label, 18))
			tabBtn.SetToolTipText(label)
			_ = tabBtn.SetMinMaxSize(walk.Size{Width: 0, Height: 0}, walk.Size{Width: measureTabTextWidth(label), Height: 0})
			if active {
				tabBtn.SetEnabled(false)
			} else {
				tabBtn.Clicked().Attach(func() { a.activateTab(idx) })
			}

			closeBtn, err := newMDL2ToolButton(wrap, glyphClose, i18n.T(i18n.KeyCloseTab), func() {
				a.closeTab(idx)
			})
			if err != nil {
				wrap.Dispose()
				continue
			}
			_ = closeBtn.SetMinMaxSize(walk.Size{Width: 22, Height: 22}, walk.Size{Width: 22, Height: 22})

			setTabCompositeActive(wrap, active)
		}

		addBtn, err := newMDL2ToolButton(a.tabBar, glyphAdd, i18n.T(i18n.KeyNewTabConnect), a.onNewTab)
		if err == nil {
			_ = addBtn.SetMinMaxSize(walk.Size{Width: 28, Height: 28}, walk.Size{Width: 28, Height: 28})
		}

		a.tabBar.RequestLayout()
	})
}
