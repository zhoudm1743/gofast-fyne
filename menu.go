package gofastfyne

import "fyne.io/fyne/v2"

// buildMainMenu 由配置构建主窗口菜单栏。
// onFire 在叶子项被点击时回调，由 Gui 注入（派发框架事件或执行内置动作）。
func buildMainMenu(items []MenuItem, onFire func(MenuItem)) *fyne.MainMenu {
	menus := make([]*fyne.Menu, 0, len(items))
	for _, item := range items {
		menus = append(menus, fyne.NewMenu(item.Label, buildMenuItems(item.Items, onFire)...))
	}
	return fyne.NewMainMenu(menus...)
}

// buildMenu 由配置构建普通菜单（托盘菜单）。
func buildMenu(items []MenuItem, onFire func(MenuItem)) *fyne.Menu {
	return fyne.NewMenu("", buildMenuItems(items, onFire)...)
}

// buildMenuItems 递归构建菜单项；叶子项点击触发 onFire。
func buildMenuItems(items []MenuItem, onFire func(MenuItem)) []*fyne.MenuItem {
	fyneItems := make([]*fyne.MenuItem, 0, len(items))
	for _, item := range items {
		if item.Separator {
			fyneItems = append(fyneItems, fyne.NewMenuItemSeparator())
			continue
		}
		mi := fyne.NewMenuItem(item.Label, func() {
			onFire(item)
		})
		if len(item.Items) > 0 {
			mi.ChildMenu = fyne.NewMenu("", buildMenuItems(item.Items, onFire)...)
		}
		fyneItems = append(fyneItems, mi)
	}
	return fyneItems
}
