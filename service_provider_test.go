package gofastfyne

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/zhoudm1743/go-fast-framework/config"
	"github.com/zhoudm1743/go-fast-framework/event"
	"github.com/zhoudm1743/go-fast-framework/foundation"
)

// useTestDriver 将 Provider 的 Fyne 应用工厂替换为无头测试驱动，
// 使 Provider 冒烟测试不依赖真实显示服务器与桌面驱动。
func useTestDriver(t *testing.T) {
	t.Helper()
	orig := newFyneApp
	newFyneApp = func(string) fyne.App { return test.NewApp() }
	t.Cleanup(func() { newFyneApp = orig })
}

func writeFile(t *testing.T, dir, name string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// newBootedApp 引导一个含配置 + 事件 + 本插件的最小应用。
func newBootedApp(t *testing.T, yaml string, extra ...foundation.ServiceProvider) foundation.Application {
	t.Helper()
	dir := t.TempDir()
	if yaml != "" {
		writeFile(t, dir, "config/config.yaml", []byte(yaml))
	}
	providers := []foundation.ServiceProvider{
		&config.ServiceProvider{},
		&event.ServiceProvider{},
		&ServiceProvider{},
	}
	providers = append(providers, extra...)
	app := foundation.NewApplication(dir)
	app.SetProviders(providers)
	app.Boot()
	return app
}

func TestServiceProvider_ConfigDefaults(t *testing.T) {
	sp := &ServiceProvider{}
	cp, ok := interface{}(sp).(foundation.ConfigProvider)
	if !ok {
		t.Fatal("ServiceProvider 应实现 ConfigProvider")
	}
	defaults := cp.ConfigDefaults()
	for _, key := range []string{"fyne.app_id", "fyne.window.title", "fyne.window.width", "fyne.window.height", "fyne.window.close"} {
		if _, exists := defaults[key]; !exists {
			t.Errorf("默认配置缺少 %s", key)
		}
	}
}

func TestServiceProvider_IsDeferred(t *testing.T) {
	sp := &ServiceProvider{}
	dp, ok := interface{}(sp).(foundation.DeferredProvider)
	if !ok {
		t.Fatal("ServiceProvider 应实现 DeferredProvider（CLI 场景不初始化 GUI）")
	}
	svcs := dp.DeferredServices()
	if len(svcs) != 1 || svcs[0] != "gui" {
		t.Errorf("DeferredServices 应返回 [gui]，实际 %v", svcs)
	}
}

func TestServiceProvider_BootBindsGui(t *testing.T) {
	useTestDriver(t)
	app := newBootedApp(t, "fyne:\n  window:\n    title: 冒烟窗口\n")

	if !app.Bound("gui") {
		t.Fatal("Boot 后 gui 服务应可解析（延迟绑定）")
	}

	gui := app.MustMake("gui").(Gui)
	if gui.Window().Title() != "冒烟窗口" {
		t.Errorf("窗口标题 = %q, 期望 冒烟窗口", gui.Window().Title())
	}
}

func TestServiceProvider_WithoutEventService(t *testing.T) {
	useTestDriver(t)
	// 无 event Provider：gui 仍可用，仅 Emit 报错。
	dir := t.TempDir()
	app := foundation.NewApplication(dir)
	app.SetProviders([]foundation.ServiceProvider{
		&config.ServiceProvider{},
		&ServiceProvider{},
	})
	app.Boot()

	gui := app.MustMake("gui").(Gui)
	if gui.Window() == nil {
		t.Fatal("窗口应已创建")
	}
	if err := gui.Emit("x", nil); err == nil {
		t.Error("无事件总线时 Emit 应返回错误")
	}
}

func TestServiceProvider_ShutdownHookQuitsGui(t *testing.T) {
	useTestDriver(t)
	app := newBootedApp(t, "")
	gui := app.MustMake("gui").(Gui)
	// 框架 Shutdown 不应 panic（内部调用 Fyne Quit）。
	app.Shutdown()
	_ = gui
}
