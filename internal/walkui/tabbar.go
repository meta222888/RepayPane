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
		_ = a.tabBar.SetMinMaxSize(walk.Size{Width: 0, Height: 0}, walk.Size{Width: 0, Height: tabBarRowHeight})

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
			hl.SetMargins(walk.Margins{})
			hl.SetSpacing(0)
			if err := wrap.SetLayout(hl); err != nil {
				wrap.Dispose()
				continue
			}
			_ = wrap.SetMinMaxSize(walk.Size{Width: 0, Height: 0}, walk.Size{Width: tabCompositeMaxWidth().Width, Height: tabBarRowHeight})

			tabBtn, err := walk.NewPushButton(wrap)
			if err != nil {
				wrap.Dispose()
				continue
			}
			tabBtn.SetText(truncateRunes(label, 18))
			tabBtn.SetToolTipText(label)
			tabW := measureTabTextWidth(truncateRunes(label, 18))
			_ = tabBtn.SetMinMaxSize(
				walk.Size{Width: tabW, Height: tabBarRowHeight - 4},
				walk.Size{Width: tabW, Height: tabBarRowHeight - 2},
			)
			if active {
				tabBtn.SetEnabled(false)
			} else {
				tabBtn.Clicked().Attach(func() { a.activateTab(idx) })
			}

			_, err = newMDL2ToolButton(wrap, glyphClose, i18n.T(i18n.KeyCloseTab), func() {
				a.closeTab(idx)
			})
			if err != nil {
				wrap.Dispose()
				continue
			}

			setTabCompositeActive(wrap, active)
		}

		_, _ = newMDL2ToolButton(a.tabBar, glyphAdd, i18n.T(i18n.KeyNewTabConnect), a.onNewTab)

		a.tabBar.RequestLayout()
	})
}
