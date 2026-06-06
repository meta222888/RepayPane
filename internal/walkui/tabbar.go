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
		a.tabBar.Children().Clear()
		for i, tab := range a.tabs {
			idx := i
			tabBtn, _ := walk.NewPushButton(a.tabBar)
			label := tab.tabLabel()
			switch tab.state {
			case tabConnecting:
				label += " …"
			case tabDisconnected:
				if tab.client == nil && tab.state == tabDisconnected {
					label += " !"
				}
			}
			tabBtn.SetText(label)
			if i == a.activeTab {
				tabBtn.SetEnabled(false)
			} else {
				tabBtn.Clicked().Attach(func() { a.activateTab(idx) })
			}
			closeBtn, _ := walk.NewPushButton(a.tabBar)
			closeBtn.SetText("×")
			closeBtn.SetMinMaxSize(walk.Size{Width: 28, Height: 0}, walk.Size{Width: 28, Height: 0})
			closeBtn.Clicked().Attach(func() { a.closeTab(idx) })
		}
		addBtn, _ := walk.NewPushButton(a.tabBar)
		addBtn.SetText(i18n.T(i18n.KeyNewTabConnect))
		addBtn.Clicked().Attach(a.onNewTab)
	})
}
