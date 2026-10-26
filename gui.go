package gofastfyne

import (
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/zhoudm1743/go-fast-framework/contracts"
)

// 主窗口名称：窗口级事件 payload 中 "window" 键的值用它标识主窗口。
const mainWindowName = "main"

// 窗口级事件名称，经框架事件总线派发，可用 On 订阅（回调在 UI 线程执行）。
// 事件 payload 携带 "window" 键标识来源窗口（主窗口为 "main"）。
const (
	EventWindowShown  = "window.shown"
	EventWindowHidden = "window.hidden"
	EventWindowClosed = "window.closed"
)

// Gui 桌面 GUI 服务契约。
// 通过 app.MustMake("gui").(gofastfyne.Gui) 获取。
type Gui interface {
	// Run 显示主窗口并阻塞运行 Fyne 事件循环；
	// 窗口退出后触发框架 Shutdown（幂等，可重复调用）。
	Run() error
	// Quit 退出 Fyne 应用（幂等）。
	Quit()

	// Window 返回主窗口。
	Window() fyne.Window
	// NewWindow 创建新窗口（不自动显示）。
	NewWindow(title string) fyne.Window
	// SetContent 设置主窗口内容。
	SetContent(c fyne.CanvasObject)
	// ShowMainWindow / HideMainWindow / ToggleMainWindow 控制主窗口显隐。
	ShowMainWindow()
	HideMainWindow()
	ToggleMainWindow()
	// Visible 返回主窗口当前是否可见。
	Visible() bool

	// App 返回底层 Fyne 应用实例（高级用法：直接访问 Settings/Preferences 等）。
	App() fyne.App

	// SetMenu 以编程方式重建主窗口菜单栏（传入 nil 清除）。
	// 与配置 fyne.menu 等效，支持运行期动态更新（未读数、最近文件等）。
	SetMenu(items []MenuItem)
	// SetTrayMenu 以编程方式重建托盘菜单；驱动不支持托盘时静默忽略。
	SetTrayMenu(items []MenuItem)

	// OpenWindow 打开（或复用）一个命名次级窗口并显示。
	// 同名窗口已存在时直接显示并返回原实例；窗口关闭时自动注销。
	OpenWindow(name, title string, content fyne.CanvasObject) fyne.Window
	// WindowByName 按名称查找次级窗口（未找到返回 nil）。
	WindowByName(name string) fyne.Window
	// CloseWindow 关闭并注销命名窗口（不存在时无操作）。
	CloseWindow(name string)

	// Emit 向框架事件总线派发字符串命名的 UI 事件，
	// 所有 On 监听者（含菜单/托盘触发路径）都会收到。
	Emit(name string, payload map[string]any) error
	// On 监听字符串命名事件，回调保证在 Fyne UI 线程执行。
	On(name string, fn func(map[string]any))
	// Listen 监听类型化的框架事件（如 OrderShipped），回调在 UI 线程执行。
	Listen(e contracts.Eventer, fn func(args ...any))
}

// 主窗口尺寸的偏好存储键。
const (
	prefWidthKey  = "gofastfyne.window.width"
	prefHeightKey = "gofastfyne.window.height"
)

// gui 是 Gui 接口的默认实现，窗口、菜单、托盘、事件桥接均在此。
type gui struct {
	fyneApp  fyne.App
	cfg      Config
	basePath string
	evt      contracts.Event
	log      contracts.Log

	appIcon fyne.Resource
	winIcon fyne.Resource
	mainWin fyne.Window

	// 次级窗口注册表：name → window
	winMu   sync.RWMutex
	windows map[string]fyne.Window

	handlersMu sync.RWMutex
	handlers   map[string][]func(map[string]any)

	onShutdown func()

	visible      atomic.Bool
	loopEnded    atomic.Bool // 事件循环已结束（Run 返回）；此后 Quit 不再触碰 Fyne
	quitOnce     sync.Once
	shutdownOnce sync.Once
}

// New 以编程方式创建 Gui，不依赖框架服务容器（高级用法与集成测试）。
// evt 可为 nil；cfg 可用 defaultConfig 起点按字段覆盖。
func New(fapp fyne.App, cfg Config, basePath string, evt contracts.Event) (Gui, error) {
	return newGui(fapp, cfg, basePath, evt)
}

// newGui 创建 gui 实例并装配主窗口、菜单、托盘。
// evt 可为 nil（无事件总线时桥接自动禁用，Emit 返回错误）。
func newGui(fapp fyne.App, cfg Config, basePath string, evt contracts.Event) (*gui, error) {
	g := &gui{
		fyneApp:  fapp,
		cfg:      cfg,
		basePath: basePath,
		evt:      evt,
		handlers: make(map[string][]func(map[string]any)),
		windows:  make(map[string]fyne.Window),
	}
	if evt != nil {
		evt.Register(map[contracts.Eventer][]contracts.EventListener{
			&Event{}: {&bridgeListener{gui: g}},
		})
	}

	wireLifecycle(fapp.Lifecycle(), g.emitEvent)

	applyTheme(fapp, cfg.Theme)

	// 应用图标：窗口默认继承；托盘图标未配置时也回退到它
	if cfg.Icon != "" {
		res, err := loadResource(basePath, cfg.Icon)
		if err != nil {
			return nil, err
		}
		fapp.SetIcon(res)
		g.appIcon = res
	}

	// 窗口尺寸记忆：偏好存储覆盖配置值
	if cfg.Window.RememberSize {
		if w := fapp.Preferences().Int(prefWidthKey); w > 0 {
			g.cfg.Window.Width = w
		}
		if h := fapp.Preferences().Int(prefHeightKey); h > 0 {
			g.cfg.Window.Height = h
		}
	}

	g.mainWin = fapp.NewWindow(g.cfg.Window.Title)
	g.mainWin.Resize(fyne.NewSize(float32(g.cfg.Window.Width), float32(g.cfg.Window.Height)))
	g.mainWin.SetFixedSize(g.cfg.Window.Fixed)
	if g.cfg.Window.Icon != "" {
		res, err := loadResource(basePath, g.cfg.Window.Icon)
		if err != nil {
			return nil, err
		}
		g.mainWin.SetIcon(res)
		g.winIcon = res
	}
	if g.cfg.Window.Close == CloseHide {
		g.mainWin.SetCloseIntercept(func() { g.HideMainWindow() })
	}
	g.mainWin.SetOnClosed(func() {
		g.saveWindowSize()
		g.emitWindowEvent(EventWindowClosed, mainWindowName)
	})
	if len(g.cfg.Menu) > 0 {
		g.mainWin.SetMainMenu(buildMainMenu(g.cfg.Menu, g.handleMenuItem))
	}
	if g.cfg.Tray.Enabled {
		if err := g.setupTray(); err != nil {
			return nil, err
		}
	}
	return g, nil
}

// setupTray 按配置装配系统托盘。非桌面驱动（如测试驱动）时静默跳过。
func (g *gui) setupTray() error {
	res, err := loadResource(g.basePath, g.cfg.Tray.Icon)
	if err != nil {
		return err
	}
	if res == nil {
		res = g.appIcon // 未配置托盘图标时回退应用图标
	}
	desk, ok := g.fyneApp.(desktop.App)
	if !ok {
		return nil
	}
	if res != nil {
		desk.SetSystemTrayIcon(res)
	}
	desk.SetSystemTrayWindow(g.mainWin)
	if len(g.cfg.Tray.Menu) > 0 {
		desk.SetSystemTrayMenu(buildMenu(g.cfg.Tray.Menu, g.handleMenuItem))
	}
	return nil
}

// handleMenuItem 菜单/托盘点击入口：Event 派发框架事件，Action 执行内置动作。
func (g *gui) handleMenuItem(item MenuItem) {
	if item.Event != "" {
		if err := g.Emit(item.Event, item.Payload); err != nil && g.log != nil {
			g.log.Errorf("[GoFast] GUI 派发事件 %s 失败: %v", item.Event, err)
		}
		return
	}
	switch item.Action {
	case ActionShow:
		g.ShowMainWindow()
	case ActionHide:
		g.HideMainWindow()
	case ActionToggle:
		g.ToggleMainWindow()
	case ActionQuit:
		g.Quit()
	}
}

func (g *gui) Window() fyne.Window { return g.mainWin }

func (g *gui) NewWindow(title string) fyne.Window { return g.fyneApp.NewWindow(title) }

func (g *gui) SetContent(c fyne.CanvasObject) { g.mainWin.SetContent(c) }

func (g *gui) ShowMainWindow() {
	g.mainWin.Show()
	g.visible.Store(true)
	g.emitWindowEvent(EventWindowShown, mainWindowName)
}

func (g *gui) HideMainWindow() {
	g.mainWin.Hide()
	g.visible.Store(false)
	g.emitWindowEvent(EventWindowHidden, mainWindowName)
}

func (g *gui) ToggleMainWindow() {
	if g.visible.Load() {
		g.HideMainWindow()
	} else {
		g.ShowMainWindow()
	}
}

// Visible 返回主窗口当前是否可见。
func (g *gui) Visible() bool { return g.visible.Load() }

// App 返回底层 Fyne 应用实例。
func (g *gui) App() fyne.App { return g.fyneApp }

// saveWindowSize 将主窗口当前尺寸写入偏好存储（remember_size 启用时）。
func (g *gui) saveWindowSize() {
	if !g.cfg.Window.RememberSize {
		return
	}
	sz := g.mainWin.Canvas().Size()
	g.fyneApp.Preferences().SetInt(prefWidthKey, int(sz.Width))
	g.fyneApp.Preferences().SetInt(prefHeightKey, int(sz.Height))
}

// Run 阻塞运行事件循环，退出后触发 onShutdown 一次。
func (g *gui) Run() error {
	g.visible.Store(true)
	g.mainWin.ShowAndRun()
	g.loopEnded.Store(true)
	g.shutdownOnce.Do(func() {
		if g.onShutdown != nil {
			g.onShutdown()
		}
	})
	return nil
}

// Quit 退出 Fyne 应用（幂等）。
// 经 fyne.Do 派发，可从任意 goroutine 安全调用（如信号处理、后台任务）。
// 事件循环已结束（如 SIGTERM 由 Fyne 自行优雅退出）时直接返回，
// 避免对已销毁的 GLFW 资源二次清理。
func (g *gui) Quit() {
	g.quitOnce.Do(func() {
		if g.loopEnded.Load() {
			return
		}
		fyne.Do(func() { g.fyneApp.Quit() })
	})
}

// emitWindowEvent 派发窗口级事件，payload 携带来源窗口名。
func (g *gui) emitWindowEvent(name, windowName string) {
	if err := g.Emit(name, map[string]any{"window": windowName}); err != nil && g.log != nil {
		g.log.Errorf("[GoFast] GUI 派发窗口事件 %s 失败: %v", name, err)
	}
}

// emitEvent 派发应用级事件（无 payload）；无事件总线或监听者时静默忽略。
func (g *gui) emitEvent(name string) {
	if err := g.Emit(name, nil); err != nil && g.log != nil {
		g.log.Errorf("[GoFast] GUI 派发事件 %s 失败: %v", name, err)
	}
}

// SetMenu 以编程方式重建主窗口菜单栏（nil 清除）。
func (g *gui) SetMenu(items []MenuItem) {
	if len(items) == 0 {
		g.mainWin.SetMainMenu(nil)
		return
	}
	g.mainWin.SetMainMenu(buildMainMenu(items, g.handleMenuItem))
}

// SetTrayMenu 以编程方式重建托盘菜单；驱动不支持托盘时静默忽略。
func (g *gui) SetTrayMenu(items []MenuItem) {
	desk, ok := g.fyneApp.(desktop.App)
	if !ok {
		if g.log != nil {
			g.log.Warnf("[GoFast] 当前 Fyne 驱动不支持系统托盘，SetTrayMenu 被忽略")
		}
		return
	}
	desk.SetSystemTrayMenu(buildMenu(items, g.handleMenuItem))
}

// OpenWindow 打开（或复用）一个命名次级窗口并显示。
func (g *gui) OpenWindow(name, title string, content fyne.CanvasObject) fyne.Window {
	g.winMu.RLock()
	w, ok := g.windows[name]
	g.winMu.RUnlock()
	if ok {
		w.Show()
		return w
	}

	w = g.fyneApp.NewWindow(title)
	w.SetContent(content)
	w.SetOnClosed(func() {
		g.winMu.Lock()
		delete(g.windows, name)
		g.winMu.Unlock()
		g.emitWindowEvent(EventWindowClosed, name)
	})
	g.winMu.Lock()
	g.windows[name] = w
	g.winMu.Unlock()
	w.Show()
	g.emitWindowEvent(EventWindowShown, name)
	return w
}

// WindowByName 按名称查找次级窗口（未找到返回 nil）。
func (g *gui) WindowByName(name string) fyne.Window {
	g.winMu.RLock()
	defer g.winMu.RUnlock()
	return g.windows[name]
}

// CloseWindow 关闭并注销命名窗口（不存在时无操作）。
func (g *gui) CloseWindow(name string) {
	g.winMu.RLock()
	w, ok := g.windows[name]
	g.winMu.RUnlock()
	if ok {
		w.Close()
	}
}

// SetOnShutdown 设置事件循环退出后的回调（Provider 注入框架 Shutdown）。
func (g *gui) SetOnShutdown(fn func()) { g.onShutdown = fn }

// SetLogger 注入框架日志服务。
func (g *gui) SetLogger(l contracts.Log) { g.log = l }

// ── 事件桥接（Emit / On / Listen）────────────────────────────────────

// On 监听字符串命名的 UI 事件（菜单/托盘/业务 Emit 均可触发）。
// 回调保证在 Fyne UI 线程执行，可安全操作界面。
func (g *gui) On(name string, fn func(map[string]any)) {
	g.handlersMu.Lock()
	defer g.handlersMu.Unlock()
	g.handlers[name] = append(g.handlers[name], fn)
}

// dispatchUI 将事件派发到匹配的处理器，回调通过 fyne.Do 切回 UI 线程。
func (g *gui) dispatchUI(name string, payload map[string]any) {
	g.handlersMu.RLock()
	fns := append([]func(map[string]any){}, g.handlers[name]...)
	g.handlersMu.RUnlock()
	if len(fns) == 0 {
		return
	}
	fyne.Do(func() {
		for _, fn := range fns {
			fn(payload)
		}
	})
}

// Emit 向框架事件总线派发字符串命名的 UI 事件，所有 On 监听者都会收到。
func (g *gui) Emit(name string, payload map[string]any) error {
	if g.evt == nil {
		return errNoEventBus
	}
	return g.evt.Job(&Event{Name: name, Payload: payload}, []contracts.EventArg{
		{Type: "string", Value: name},
		{Type: "map", Value: payload},
	}).Dispatch()
}

// Listen 监听类型化的框架事件（如 OrderShipped）。
// 回调通过 fyne.Do 切回 UI 线程执行，参数为事件载荷值。
func (g *gui) Listen(e contracts.Eventer, fn func(args ...any)) {
	if g.evt == nil {
		return
	}
	g.evt.Register(map[contracts.Eventer][]contracts.EventListener{
		e: {&uiListener{sig: "gofastfyne.listen." + eventTypeName(e), fn: fn}},
	})
}
