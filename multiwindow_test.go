package gofastfyne

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/zhoudm1743/go-fast-framework/event"
)

func TestGui_OpenWindow_RegistersByName(t *testing.T) {
	g, err := newGui(test.NewApp(), defaultConfig(), t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}

	w := g.OpenWindow("detail", "详情", widget.NewLabel("content"))
	if w == nil || w.Title() != "详情" {
		t.Fatalf("OpenWindow 应创建命名窗口: %+v", w)
	}
	if g.WindowByName("detail") != w {
		t.Error("WindowByName 应返回已注册的窗口")
	}
	if g.WindowByName("nope") != nil {
		t.Error("未注册的名称应返回 nil")
	}
}

func TestGui_OpenWindow_DedupesByName(t *testing.T) {
	g, err := newGui(test.NewApp(), defaultConfig(), t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}

	first := g.OpenWindow("detail", "详情", widget.NewLabel("a"))
	second := g.OpenWindow("detail", "另一个标题", widget.NewLabel("b"))
	if first != second {
		t.Error("同名窗口重复打开应返回同一实例")
	}
}

func TestGui_CloseWindow_RemovesFromRegistry(t *testing.T) {
	g, err := newGui(test.NewApp(), defaultConfig(), t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}

	g.OpenWindow("detail", "详情", widget.NewLabel("x"))
	g.CloseWindow("detail")

	if g.WindowByName("detail") != nil {
		t.Error("关闭后应从注册表移除")
	}
	g.CloseWindow("detail") // 重复关闭不应 panic
}

func TestGui_WindowEventPayloadIdentifiesWindow(t *testing.T) {
	g, err := newGui(test.NewApp(), defaultConfig(), t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}

	type evt struct {
		name   string
		window string
	}
	var events []evt
	record := func(n string) func(map[string]any) {
		return func(p map[string]any) {
			win, _ := p["window"].(string)
			events = append(events, evt{n, win})
		}
	}
	g.On(EventWindowShown, record(EventWindowShown))
	g.On(EventWindowClosed, record(EventWindowClosed))

	g.ShowMainWindow() // 主窗口
	g.OpenWindow("detail", "详情", widget.NewLabel("x"))
	g.CloseWindow("detail")

	if len(events) != 3 {
		t.Fatalf("应有 3 个窗口事件，实际 %v", events)
	}
	if events[0] != (evt{EventWindowShown, "main"}) {
		t.Errorf("主窗口 shown 事件错误: %+v", events[0])
	}
	if events[1] != (evt{EventWindowShown, "detail"}) {
		t.Errorf("次级窗口 shown 事件错误: %+v", events[1])
	}
	if events[2] != (evt{EventWindowClosed, "detail"}) {
		t.Errorf("次级窗口 closed 事件错误: %+v", events[2])
	}
}
