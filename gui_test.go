package gofastfyne

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/zhoudm1743/go-fast-framework/event"
)

// newTestGuiWithConfig 用测试驱动创建 gui，返回配置便于断言。
func newTestGuiWithConfig(t *testing.T, cfg Config) *gui {
	t.Helper()
	g, err := newGui(test.NewApp(), cfg, t.TempDir(), event.New())
	if err != nil {
		t.Fatalf("创建 Gui 失败: %v", err)
	}
	return g
}

func TestGui_MainWindowFromConfig(t *testing.T) {
	cfg := defaultConfig()
	cfg.Window.Title = "测试窗口"
	cfg.Window.Width = 800
	cfg.Window.Height = 600
	cfg.Window.Fixed = true

	g := newTestGuiWithConfig(t, cfg)
	win := g.Window()

	if win.Title() != "测试窗口" {
		t.Errorf("窗口标题 = %q, 期望 测试窗口", win.Title())
	}
	if win.FixedSize() != true {
		t.Error("FixedSize 应为 true")
	}
	if sz := win.Canvas().Size(); sz.Width != 800 || sz.Height != 600 {
		t.Errorf("窗口尺寸 = %v, 期望 800x600", sz)
	}
}

func TestGui_MainMenuApplied(t *testing.T) {
	cfg := defaultConfig()
	cfg.Menu = []MenuItem{
		{Label: "文件", Items: []MenuItem{{Label: "打开", Event: "file.open"}}},
		{Label: "帮助", Items: []MenuItem{{Label: "关于", Event: "app.about"}}},
	}

	g := newTestGuiWithConfig(t, cfg)
	main := g.Window().MainMenu()
	if main == nil {
		t.Fatal("应设置主菜单")
	}
	if len(main.Items) != 2 || main.Items[0].Label != "文件" {
		t.Errorf("主菜单结构错误: %+v", main.Items)
	}
}

func TestGui_MenuClickDispatchesEvent(t *testing.T) {
	cfg := defaultConfig()
	cfg.Menu = []MenuItem{
		{Label: "文件", Items: []MenuItem{
			{Label: "打开", Event: "file.open", Payload: map[string]any{"target": "recent"}},
		}},
	}

	g := newTestGuiWithConfig(t, cfg)

	var got map[string]any
	g.On("file.open", func(p map[string]any) { got = p })

	item := g.Window().MainMenu().Items[0].Items[0]
	item.Action()

	if got == nil || got["target"] != "recent" {
		t.Errorf("菜单点击应派发事件到处理器，实际 %+v", got)
	}
}

func TestGui_MenuClickBuiltInShowAction(t *testing.T) {
	cfg := defaultConfig()
	cfg.Menu = []MenuItem{
		{Label: "视图", Items: []MenuItem{
			{Label: "显示主窗口", Action: ActionShow},
			{Label: "隐藏主窗口", Action: ActionHide},
			{Label: "切换", Action: ActionToggle},
		}},
	}

	g := newTestGuiWithConfig(t, cfg)
	items := g.Window().MainMenu().Items[0].Items

	g.HideMainWindow()
	if g.Visible() {
		t.Fatal("HideMainWindow 后窗口应隐藏")
	}
	items[0].Action() // show
	if !g.Visible() {
		t.Error("点击显示后窗口应可见")
	}
	items[1].Action() // hide
	if g.Visible() {
		t.Error("点击隐藏后窗口应隐藏")
	}
	items[2].Action() // toggle
	if !g.Visible() {
		t.Error("toggle 应恢复显示")
	}
}

func TestGui_SetContent(t *testing.T) {
	g := newTestGuiWithConfig(t, defaultConfig())
	g.SetContent(widget.NewLabel("hello"))
	if g.Window().Content() == nil {
		t.Error("SetContent 后窗口应有内容")
	}
}

func TestGui_NewWindow(t *testing.T) {
	g := newTestGuiWithConfig(t, defaultConfig())
	win := g.NewWindow("子窗口")
	if win == nil || win.Title() != "子窗口" {
		t.Errorf("NewWindow 应返回命名窗口: %+v", win)
	}
}

func TestGui_TrayEnabledWithIcon(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "icon.png", []byte{0x89, 0x50, 0x4E, 0x47})

	cfg := defaultConfig()
	cfg.Tray.Enabled = true
	cfg.Tray.Icon = "icon.png"
	cfg.Tray.Menu = []MenuItem{{Label: "退出", Action: ActionQuit}}

	if _, err := newGui(test.NewApp(), cfg, dir, event.New()); err != nil {
		t.Errorf("托盘启用且图标存在时应成功: %v", err)
	}
}

func TestGui_TrayIconMissingErrors(t *testing.T) {
	cfg := defaultConfig()
	cfg.Tray.Enabled = true
	cfg.Tray.Icon = "nope.png"

	if _, err := newGui(test.NewApp(), cfg, t.TempDir(), event.New()); err == nil {
		t.Error("托盘图标缺失时 newGui 应返回错误")
	}
}

func TestGui_RunTriggersShutdownOnce(t *testing.T) {
	g := newTestGuiWithConfig(t, defaultConfig())

	calls := 0
	g.SetOnShutdown(func() { calls++ })

	if err := g.Run(); err != nil {
		t.Fatalf("Run 失败: %v", err)
	}
	_ = g.Run()
	if calls != 1 {
		t.Errorf("退出回调应只执行一次，实际 %d 次", calls)
	}
}

func TestGui_AppAndSizeContract(t *testing.T) {
	// Gui 接口完整性：确保 gui 实现公开接口，便于 Provider 暴露。
	var _ Gui = (*gui)(nil)
	var _ fyne.CanvasObject = widget.NewLabel("x")
}
