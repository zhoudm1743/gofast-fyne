package gofastfyne

// 应用级生命周期事件名称，经框架事件总线派发，可用 On 订阅。
// 注意：app.stopped 触发时事件循环正在收尾，fyne.Do 派发的 UI 更新可能来不及执行，
// 重量级清理请使用框架 Shutdown 钩子（app.OnShutdown）。
const (
	EventAppStarted    = "app.started"    // 应用进入运行（事件循环启动）
	EventAppStopped    = "app.stopped"    // 应用停止运行（事件循环结束）
	EventAppForeground = "app.foreground" // 应用进入前台获得焦点
	EventAppBackground = "app.background" // 应用退入后台
)

// lifecycleTarget 是 fyne.Lifecycle 的最小需求面，便于测试替身。
type lifecycleTarget interface {
	SetOnStarted(func())
	SetOnStopped(func())
	SetOnEnteredForeground(func())
	SetOnExitedForeground(func())
}

// wireLifecycle 把 Fyne 应用生命周期转成框架事件总线上的字符串事件。
func wireLifecycle(l lifecycleTarget, emit func(string)) {
	l.SetOnStarted(func() { emit(EventAppStarted) })
	l.SetOnStopped(func() { emit(EventAppStopped) })
	l.SetOnEnteredForeground(func() { emit(EventAppForeground) })
	l.SetOnExitedForeground(func() { emit(EventAppBackground) })
}
