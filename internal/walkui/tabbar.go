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
		_ = a.tabBar.SetMinMaxSize(walk.Size{Width: 0, Height: 0}, walk.Size{Width: 0, Height: tabBarHeight})

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
			hl.SetSpacing(2)
			if err := wrap.SetLayout(hl); err != nil {
				wrap.Dispose()
				continue
			}
			_ = wrap.SetMinMaxSize(walk.Size{Width: 0, Height: 0}, walk.Size{Width: tabMaxWidth, Height: tabBarHeight})

			dot, err := walk.NewLabel(wrap)
			if err == nil {
				dot.SetText("●")
				dot.SetTextColor(tabStateColor(tab.state))
				dotFont, _ := walk.NewFont("Segoe UI", 6, walk.FontBold)
				if dotFont != nil {
					dot.SetFont(dotFont)
				}
				_ = dot.SetMinMaxSize(walk.Size{Width: statusDotTab, Height: statusDotTab}, walk.Size{Width: statusDotTab, Height: statusDotTab})
			}

			tabBtn, err := walk.NewPushButton(wrap)
			if err != nil {
				wrap.Dispose()
				continue
			}
			tabBtn.SetText(truncateRunes(label, 18))
			tabBtn.SetToolTipText(label)
			tabW := measureTabTextWidth(truncateRunes(label, 18))
			_ = tabBtn.SetMinMaxSize(
				walk.Size{Width: tabW, Height: tabBarHeight - 4},
				walk.Size{Width: tabW, Height: tabBarHeight - 2},
			)
			if active {
				tabBtn.SetEnabled(false)
			} else {
				tabBtn.Clicked().Attach(func() { a.activateTab(idx) })
			}

			_, err = newPNGToolButton(wrap, UIBmpClose(), i18n.T(i18n.KeyCloseTab), func() {
				a.closeTab(idx)
			})
			if err != nil {
				wrap.Dispose()
				continue
			}

			setTabCompositeActive(wrap, active)
		}

		_, _ = newPNGToolButton(a.tabBar, UIBmpNew(), i18n.T(i18n.KeyNewTabConnect), a.onNewTab)

		a.tabBar.RequestLayout()
	})
}
