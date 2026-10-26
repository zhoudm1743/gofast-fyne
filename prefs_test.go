package gofastfyne

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/zhoudm1743/go-fast-framework/event"
)

func TestGui_RememberSize_LoadsFromPreferences(t *testing.T) {
	a := test.NewApp()
	a.Preferences().SetInt(prefWidthKey, 640)
	a.Preferences().SetInt(prefHeightKey, 480)

	cfg := defaultConfig()
	cfg.Window.RememberSize = true

	g, err := newGui(a, cfg, t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}
	if sz := g.Window().Canvas().Size(); sz.Width != 640 || sz.Height != 480 {
		t.Errorf("窗口尺寸 = %v, 期望从偏好恢复 640x480", sz)
	}
}

func TestGui_RememberSize_SavesOnClose(t *testing.T) {
	a := test.NewApp()
	cfg := defaultConfig()
	cfg.Window.RememberSize = true
	cfg.Window.Width = 800
	cfg.Window.Height = 600

	g, err := newGui(a, cfg, t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}
	g.Window().Close()

	if got := a.Preferences().Int(prefWidthKey); got != 800 {
		t.Errorf("偏好宽度 = %d, 期望 800", got)
	}
	if got := a.Preferences().Int(prefHeightKey); got != 600 {
		t.Errorf("偏好高度 = %d, 期望 600", got)
	}
}

func TestGui_RememberSize_DisabledIgnoresPreferences(t *testing.T) {
	a := test.NewApp()
	a.Preferences().SetInt(prefWidthKey, 640)
	a.Preferences().SetInt(prefHeightKey, 480)

	cfg := defaultConfig() // RememberSize 默认 false
	g, err := newGui(a, cfg, t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}
	if sz := g.Window().Canvas().Size(); sz.Width != 1024 || sz.Height != 768 {
		t.Errorf("未启用记忆时应使用配置尺寸 1024x768，实际 %v", sz)
	}
}
