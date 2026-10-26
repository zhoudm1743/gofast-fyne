package gofastfyne

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/zhoudm1743/go-fast-framework/contracts"
	"github.com/zhoudm1743/go-fast-framework/event"
)

// demoEvent 测试用的类型化框架事件。
type demoEvent struct{}

func (demoEvent) Handle(args []contracts.EventArg) ([]contracts.EventArg, error) { return args, nil }

func newTestGui(t *testing.T, evt contracts.Event) *gui {
	t.Helper()
	g, err := newGui(test.NewApp(), defaultConfig(), t.TempDir(), evt)
	if err != nil {
		t.Fatalf("创建 Gui 失败: %v", err)
	}
	return g
}

func TestGui_EmitFiresOnHandler(t *testing.T) {
	evt := event.New()
	g := newTestGui(t, evt)

	var gotPayload map[string]any
	g.On("file.open", func(p map[string]any) {
		gotPayload = p
	})

	if err := g.Emit("file.open", map[string]any{"target": "recent"}); err != nil {
		t.Fatalf("Emit 失败: %v", err)
	}

	if gotPayload == nil || gotPayload["target"] != "recent" {
		t.Errorf("处理器应收到 payload，实际 %+v", gotPayload)
	}
}

func TestGui_OnIgnoresOtherEventNames(t *testing.T) {
	evt := event.New()
	g := newTestGui(t, evt)

	calls := 0
	g.On("file.open", func(map[string]any) { calls++ })

	if err := g.Emit("other.event", nil); err != nil {
		t.Fatalf("Emit 失败: %v", err)
	}
	if calls != 0 {
		t.Errorf("不应触发的处理器被调用了 %d 次", calls)
	}
}

func TestGui_EmitMultipleHandlers(t *testing.T) {
	evt := event.New()
	g := newTestGui(t, evt)

	calls := 0
	g.On("app.quit", func(map[string]any) { calls++ })
	g.On("app.quit", func(map[string]any) { calls++ })
	g.On("app.quit", func(map[string]any) { calls++ })

	if err := g.Emit("app.quit", nil); err != nil {
		t.Fatalf("Emit 失败: %v", err)
	}
	if calls != 3 {
		t.Errorf("同名事件应有 3 个处理器执行，实际 %d", calls)
	}
}

func TestGui_ListenTypedFrameworkEvent(t *testing.T) {
	evt := event.New()
	g := newTestGui(t, evt)

	var got []any
	g.Listen(&demoEvent{}, func(args ...any) { got = args })

	args := []contracts.EventArg{{Type: "string", Value: "order-42"}, {Type: "int", Value: 7}}
	if err := evt.Job(&demoEvent{}, args).Dispatch(); err != nil {
		t.Fatalf("派发事件失败: %v", err)
	}

	if len(got) != 2 || got[0] != "order-42" || got[1] != 7 {
		t.Errorf("类型化事件参数透传错误: %+v", got)
	}
}

func TestGui_EmitWithoutEventBusErrors(t *testing.T) {
	g := newTestGui(t, nil)
	if err := g.Emit("x", nil); err == nil {
		t.Error("无事件总线时 Emit 应返回错误")
	}
}
