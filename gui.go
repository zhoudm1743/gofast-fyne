package gofastfyne

import (
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/zhoudm1743/go-fast-framework/contracts"
)

// 窗口级事件名称，经框架事件总线派发，可用 On 订阅（回调在 UI 线程执行）。
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

	// Emit 向框架事件总线派发字符串命名的 UI 事件，
	// 所有 On 监听者（含菜单/托盘触发路径）都会收到。
	Emit(name string, payload map[string]any) error
	// On 监听字符串命名事件，回调保证在 Fyne UI 线程执行。
	On(name string, fn func(map[string]any))
	// Listen 监听类型化的框架事件（如 OrderShipped），回调在 UI 线程执行。
	Listen(e contracts.Eventer, fn func(args ...any))
}

// gui 是 Gui 接口的默认实现，窗口、菜单、托盘、事件桥接均在此。
type gui struct {
	fyneApp  fyne.App
	cfg      Config
	basePath string
	evt      contracts.Event
	log      contracts.Log

	mainWin fyne.Window

	handlersMu sync.RWMutex
	handlers   map[string][]func(map[string]any)

	onShutdown func()

	visible      atomic.Bool
	quitOnce     sync.Once
	shutdownOnce sync.Once
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
	}
	if evt != nil {
		evt.Register(map[contracts.Eventer][]contracts.EventListener{
			&Event{}: {&bridgeListener{gui: g}},
		})
	}

	g.mainWin = fapp.NewWindow(cfg.Window.Title)
	g.mainWin.Resize(fyne.NewSize(float32(cfg.Window.Width), float32(cfg.Window.Height)))
	g.mainWin.SetFixedSize(cfg.Window.Fixed)
	if cfg.Window.Close == CloseHide {
		g.mainWin.SetCloseIntercept(func() { g.HideMainWindow() })
	}
	g.mainWin.SetOnClosed(func() { g.emitEvent(EventWindowClosed) })
	if len(cfg.Menu) > 0 {
		g.mainWin.SetMainMenu(buildMainMenu(cfg.Menu, g.handleMenuItem))
	}
	if cfg.Tray.Enabled {
		if err := g.setupTray(); err != nil {
			return nil, err
		}
	}
	return g, nil
}

// setupTray 按配置装配系统托盘。非桌面驱动（如测试驱动）时静默跳过。
func (g *gui) setupTray() error {
	res, err := loadTrayIcon(g.basePath, g.cfg.Tray.Icon)
	if err != nil {
		return err
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
	g.emitEvent(EventWindowShown)
}

func (g *gui) HideMainWindow() {
	g.mainWin.Hide()
	g.visible.Store(false)
	g.emitEvent(EventWindowHidden)
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

// Run 阻塞运行事件循环，退出后触发 onShutdown 一次。
func (g *gui) Run() error {
	g.visible.Store(true)
	g.mainWin.ShowAndRun()
	g.shutdownOnce.Do(func() {
		if g.onShutdown != nil {
			g.onShutdown()
		}
	})
	return nil
}

func (g *gui) Quit() {
	g.quitOnce.Do(func() { g.fyneApp.Quit() })
}

// emitEvent 派发窗口级事件；无事件总线或监听者时静默忽略。
func (g *gui) emitEvent(name string) {
	if err := g.Emit(name, nil); err != nil && g.log != nil {
		g.log.Errorf("[GoFast] GUI 派发窗口事件 %s 失败: %v", name, err)
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
