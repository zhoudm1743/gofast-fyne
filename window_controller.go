package gofastfyne

import "fyne.io/fyne/v2"

// WindowController 窗口控制器契约（类比框架 HTTP 的 contracts.Controller）。
// 每个控制器负责主窗口内的一个页面：View 返回页面视图，
// 并可通过 g 控制窗口、绑定菜单/窗口/业务事件。
type WindowController interface {
	// Title 页面标题，用于窗口内导航（如 AppTabs 标签页）。
	Title() string
	// View 构建页面视图；g 提供窗口控制与事件桥接能力。
	View(g Gui) fyne.CanvasObject
}
