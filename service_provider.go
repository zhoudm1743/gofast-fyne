package gofastfyne

import (
	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"github.com/zhoudm1743/go-fast-framework/contracts"
	"github.com/zhoudm1743/go-fast-framework/foundation"
)

// ServiceProvider Fyne 桌面 GUI 驱动。
// 实现 DeferredProvider：仅当首次解析 "gui" 服务时才初始化 Fyne，
// 纯 CLI / 队列 worker 场景完全不加载 GUI 依赖。
type ServiceProvider struct{}

var (
	_ foundation.ServiceProvider  = (*ServiceProvider)(nil)
	_ foundation.ConfigProvider   = (*ServiceProvider)(nil)
	_ foundation.DeferredProvider = (*ServiceProvider)(nil)
)

// newFyneApp 创建 Fyne 应用实例；测试可替换为 test.NewApp 注入测试驱动。
var newFyneApp = func(id string) fyne.App { return fyneapp.NewWithID(id) }

// Register 延迟绑定 "gui" 服务。
func (sp *ServiceProvider) Register(app foundation.Application) {
	app.Singleton("gui", func(app foundation.Application) (any, error) {
		cfg := loadConfig(app.Config())

		var evt contracts.Event
		if app.Bound("event") {
			evt = app.MustMake("event").(contracts.Event)
		}

		g, err := newGui(newFyneApp(cfg.AppID), cfg, app.BasePath(), evt)
		if err != nil {
			return nil, err
		}
		if app.Bound("log") {
			g.SetLogger(app.MustMake("log").(contracts.Log))
		}

		// 双向优雅退出：框架 Shutdown → Fyne 退出；Fyne 窗口关闭 → 框架 Shutdown（幂等）。
		app.OnShutdown(g.Quit)
		g.SetOnShutdown(app.Shutdown)
		return g, nil
	})
}

// Boot 无即时初始化工作（延迟到首次 Make("gui")）。
func (sp *ServiceProvider) Boot(app foundation.Application) error { return nil }

// ConfigDefaults 插件默认配置（实现 foundation.ConfigProvider）。
func (sp *ServiceProvider) ConfigDefaults() map[string]any {
	d := defaultConfig()
	return map[string]any{
		"fyne.app_id":        d.AppID,
		"fyne.window.title":  d.Window.Title,
		"fyne.window.width":  d.Window.Width,
		"fyne.window.height": d.Window.Height,
		"fyne.window.fixed":  d.Window.Fixed,
		"fyne.window.close":  d.Window.Close,
		"fyne.tray.enabled":  d.Tray.Enabled,
	}
}

// DeferredServices 声明延迟提供的服务 key（实现 foundation.DeferredProvider）。
func (sp *ServiceProvider) DeferredServices() []string { return []string{"gui"} }
