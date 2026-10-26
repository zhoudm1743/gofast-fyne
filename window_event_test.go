package gofastfyne

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/zhoudm1743/go-fast-framework/event"
)

// collectWindowEvents 订阅全部窗口级事件并记录触发的名称。
func collectWindowEvents(g *gui) *[]string {
	names := &[]string{}
	g.On(EventWindowShown, func(map[string]any) { *names = append(*names, EventWindowShown) })
	g.On(EventWindowHidden, func(map[string]any) { *names = append(*names, EventWindowHidden) })
	g.On(EventWindowClosed, func(map[string]any) { *names = append(*names, EventWindowClosed) })
	return names
}

func TestGui_ShowMainWindowEmitsShown(t *testing.T) {
	g, err := newGui(test.NewApp(), defaultConfig(), t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}
	names := collectWindowEvents(g)

	g.ShowMainWindow()

	if len(*names) != 1 || (*names)[0] != EventWindowShown {
		t.Errorf("ShowMainWindow 应触发 window.shown，实际 %v", *names)
	}
}

func TestGui_HideMainWindowEmitsHidden(t *testing.T) {
	g, err := newGui(test.NewApp(), defaultConfig(), t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}
	names := collectWindowEvents(g)

	g.HideMainWindow()

	if len(*names) != 1 || (*names)[0] != EventWindowHidden {
		t.Errorf("HideMainWindow 应触发 window.hidden，实际 %v", *names)
	}
}

func TestGui_ToggleEmitsShownAndHidden(t *testing.T) {
	g, err := newGui(test.NewApp(), defaultConfig(), t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}
	names := collectWindowEvents(g)

	g.ShowMainWindow()
	g.ToggleMainWindow() // → hidden
	g.ToggleMainWindow() // → shown

	want := []string{EventWindowShown, EventWindowHidden, EventWindowShown}
	if len(*names) != len(want) {
		t.Fatalf("事件序列 = %v, 期望 %v", *names, want)
	}
	for i := range want {
		if (*names)[i] != want[i] {
			t.Fatalf("事件序列 = %v, 期望 %v", *names, want)
		}
	}
}

func TestGui_WindowCloseEmitsClosed(t *testing.T) {
	g, err := newGui(test.NewApp(), defaultConfig(), t.TempDir(), event.New())
	if err != nil {
		t.Fatal(err)
	}
	names := collectWindowEvents(g)

	g.Window().Close()

	found := false
	for _, n := range *names {
		if n == EventWindowClosed {
			found = true
		}
	}
	if !found {
		t.Errorf("窗口关闭应触发 window.closed，实际 %v", *names)
	}
}

// 注意：Fyne 测试驱动的 window.Close() 不经过 close intercept（直接触发 onClosed），
// 因此 close=hide 的拦截路径（隐藏而非退出）无法在无头环境断言；
// 对应实现在 gui.go 的 SetCloseIntercept 处，语义以真实驱动为准。
