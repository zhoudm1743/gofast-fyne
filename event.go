package gofastfyne

import (
	"errors"
	"fmt"
	"reflect"

	"fyne.io/fyne/v2"
	"github.com/zhoudm1743/go-fast-framework/contracts"
)

// errNoEventBus 框架 event 服务未注册时 Emit/Listen 不可用。
var errNoEventBus = errors.New("[GoFast] GUI 事件总线不可用：框架 event 服务未注册")

// Event 是 UI 侧使用的字符串命名事件，实现 contracts.Eventer。
// 它作为桥接载体：Emit 以它派发，bridgeListener 按 Name 路由到 On 注册的处理器。
type Event struct {
	Name    string
	Payload map[string]any
}

// Handle 透传事件参数，不做加工。
func (e *Event) Handle(args []contracts.EventArg) ([]contracts.EventArg, error) {
	return args, nil
}

// bridgeListener 接收框架事件总线上的 Event 并按名称分发到 UI 处理器。
type bridgeListener struct {
	gui *gui
}

func (l *bridgeListener) Signature() string { return "gofastfyne.event" }

// Queue 同步执行：UI 事件需要即时反馈。
func (l *bridgeListener) Queue(args ...any) contracts.EventQueue { return contracts.EventQueue{} }

func (l *bridgeListener) Handle(args ...any) error {
	var name string
	var payload map[string]any
	if len(args) > 0 {
		name, _ = args[0].(string)
	}
	if len(args) > 1 {
		payload, _ = args[1].(map[string]any)
	}
	l.gui.dispatchUI(name, payload)
	return nil
}

// uiListener 包装用户回调的框架监听器，保证回调在 Fyne UI 线程执行。
type uiListener struct {
	sig string
	fn  func(args ...any)
}

func (l *uiListener) Signature() string { return l.sig }

func (l *uiListener) Queue(args ...any) contracts.EventQueue { return contracts.EventQueue{} }

func (l *uiListener) Handle(args ...any) error {
	// 框架事件可能由队列 worker 等后台 goroutine 派发，
	// 回调统一通过 fyne.Do 切回 UI 线程。
	fyne.Do(func() { l.fn(args...) })
	return nil
}

func eventTypeName(e contracts.Eventer) string {
	t := reflect.TypeOf(e)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}
