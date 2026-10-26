package gofastfyne

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/zhoudm1743/go-fast-framework/event"
)

func TestGui_SetMenu_ReplacesMainMenu(t *testing.T) {
	cfg := defaultConfig()
	cfg.Menu = []MenuItem{
		{Label: "旧菜单", Items: []MenuItem{{Label: "旧项", Event: "old"}}},
	}
	g, err := newGui(test.NewApp(), cfg, t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}
	if got := g.Window().MainMenu(); got == nil || got.Items[0].Label != "旧菜单" {
		t.Fatalf("初始菜单应来自配置: %+v", got)
	}

	g.SetMenu([]MenuItem{
		{Label: "新菜单", Items: []MenuItem{{Label: "新项", Event: "new"}}},
	})
	got := g.Window().MainMenu()
	if got == nil || len(got.Items) != 1 || got.Items[0].Label != "新菜单" {
		t.Fatalf("SetMenu 应替换主菜单: %+v", got)
	}

	g.SetMenu(nil)
	if g.Window().MainMenu() != nil {
		t.Error("SetMenu(nil) 应清除主菜单")
	}
}

func TestGui_SetMenu_ClickDispatchesEvent(t *testing.T) {
	g, err := newGui(test.NewApp(), defaultConfig(), t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	g.On("new.event", func(p map[string]any) { got = p })

	g.SetMenu([]MenuItem{
		{Label: "动态", Items: []MenuItem{{Label: "触发", Event: "new.event", Payload: map[string]any{"k": 1}}}},
	})
	g.Window().MainMenu().Items[0].Items[0].Action()

	if got == nil || got["k"] != 1 {
		t.Errorf("动态菜单点击应派发事件，实际 %+v", got)
	}
}

func TestGui_SetTrayMenu_OnTestDriverNoPanic(t *testing.T) {
	g, err := newGui(test.NewApp(), defaultConfig(), t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}
	// 测试驱动非 desktop.App：应静默跳过，不 panic
	g.SetTrayMenu([]MenuItem{{Label: "退出", Action: ActionQuit}})
}
