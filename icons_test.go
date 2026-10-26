package gofastfyne

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/zhoudm1743/go-fast-framework/event"
)

// 注意：Fyne 测试驱动的 SetIcon 为空实现，图标是否落到驱动无法观测；
// 这里断言 gui 侧的加载结果（appIcon/winIcon 字段），newGui 同时会把资源设置给真实驱动。
func TestGui_AppIconApplied(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app.png", []byte{0x89, 0x50, 0x4E, 0x47})

	cfg := defaultConfig()
	cfg.Icon = "app.png"

	g, err := newGui(test.NewApp(), cfg, dir, event.New())
	if err != nil {
		t.Fatal(err)
	}
	if g.appIcon == nil {
		t.Error("应用图标应加载到 gui（并已由 newGui 设置到 Fyne App）")
	}
	// 未单独配置窗口图标时 winIcon 为 nil：真实驱动下窗口自动继承应用图标
	if g.winIcon != nil {
		t.Error("未配置窗口图标时 winIcon 应为 nil（由窗口继承应用图标）")
	}
}

func TestGui_WindowIconOverridesAppIcon(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app.png", []byte{0x89})
	writeFile(t, dir, "win.png", []byte{0x50})

	cfg := defaultConfig()
	cfg.Icon = "app.png"
	cfg.Window.Icon = "win.png"

	g, err := newGui(test.NewApp(), cfg, dir, event.New())
	if err != nil {
		t.Fatal(err)
	}
	if g.appIcon == nil || g.winIcon == nil {
		t.Fatal("应用与窗口图标都应加载")
	}
	if g.winIcon.Name() == g.appIcon.Name() {
		t.Errorf("窗口图标应覆盖应用图标，实际窗口=%q 应用=%q", g.winIcon.Name(), g.appIcon.Name())
	}
}

func TestGui_MissingIconFileErrors(t *testing.T) {
	cfg := defaultConfig()
	cfg.Icon = "nope.png"
	if _, err := newGui(test.NewApp(), cfg, t.TempDir(), event.New()); err == nil {
		t.Error("应用图标文件缺失时 newGui 应返回错误")
	}

	cfg = defaultConfig()
	cfg.Window.Icon = "nope.png"
	if _, err := newGui(test.NewApp(), cfg, t.TempDir(), event.New()); err == nil {
		t.Error("窗口图标文件缺失时 newGui 应返回错误")
	}
}

func TestGui_TrayIconFallsBackToAppIcon(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app.png", []byte{0x89})

	cfg := defaultConfig()
	cfg.Icon = "app.png"
	cfg.Tray.Enabled = true // 未配置 tray.icon → 回退应用图标

	if _, err := newGui(test.NewApp(), cfg, dir, event.New()); err != nil {
		t.Errorf("托盘图标回退应用图标时应成功: %v", err)
	}
}
