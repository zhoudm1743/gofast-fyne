package gofastfyne

import "github.com/zhoudm1743/go-fast-framework/contracts"

// 窗口关闭行为。
const (
	CloseQuit = "quit" // 关闭主窗口即退出应用（默认）
	CloseHide = "hide" // 关闭主窗口时仅隐藏（配合托盘使用）
)

// 菜单/托盘项的内置动作。
const (
	ActionShow   = "show"   // 显示主窗口
	ActionHide   = "hide"   // 隐藏主窗口
	ActionToggle = "toggle" // 切换主窗口显隐
	ActionQuit   = "quit"   // 退出应用
)

// MenuItem 菜单项配置，支持嵌套子菜单。
// Event 与 Action 二选一：Event 点击后向框架事件总线派发同名事件；
// Action 执行内置动作（show / hide / toggle / quit）。
type MenuItem struct {
	Label     string
	Event     string
	Action    string
	Separator bool
	Payload   map[string]any
	Items     []MenuItem
}

// WindowConfig 主窗口配置。
type WindowConfig struct {
	Title  string
	Width  int
	Height int
	Fixed  bool   // 固定尺寸，禁止缩放
	Close  string // CloseQuit | CloseHide
}

// TrayConfig 系统托盘配置。
type TrayConfig struct {
	Enabled bool
	Icon    string // 图标文件路径（相对应用根目录）
	Tooltip string
	Menu    []MenuItem
}

// Config 插件配置，对应 config.yaml 中的 fyne 段。
type Config struct {
	AppID  string
	Window WindowConfig
	Menu   []MenuItem
	Tray   TrayConfig
}

// defaultConfig 内置默认值；同时被 ServiceProvider.ConfigDefaults 使用。
func defaultConfig() Config {
	return Config{
		AppID: "com.gofast.desktop",
		Window: WindowConfig{
			Title:  "GoFast",
			Width:  1024,
			Height: 768,
			Close:  CloseQuit,
		},
	}
}

// loadConfig 从框架配置服务读取 fyne 段，缺失键回退到内置默认值。
func loadConfig(cfg contracts.Config) Config {
	c := defaultConfig()
	c.AppID = cfg.GetString("fyne.app_id", c.AppID)
	c.Window.Title = cfg.GetString("fyne.window.title", c.Window.Title)
	c.Window.Width = cfg.GetInt("fyne.window.width", c.Window.Width)
	c.Window.Height = cfg.GetInt("fyne.window.height", c.Window.Height)
	c.Window.Fixed = cfg.GetBool("fyne.window.fixed", c.Window.Fixed)
	c.Window.Close = normalizeClose(cfg.GetString("fyne.window.close", c.Window.Close))
	c.Menu = parseMenuItems(cfg.Get("fyne.menu"))
	c.Tray.Enabled = cfg.GetBool("fyne.tray.enabled", c.Tray.Enabled)
	c.Tray.Icon = cfg.GetString("fyne.tray.icon", c.Tray.Icon)
	c.Tray.Tooltip = cfg.GetString("fyne.tray.tooltip", c.Tray.Tooltip)
	c.Tray.Menu = parseMenuItems(cfg.Get("fyne.tray.menu"))
	return c
}

func normalizeClose(v string) string {
	if v == CloseHide {
		return CloseHide
	}
	return CloseQuit
}

// parseMenuItems 容错解析菜单树：非 map 的条目与非数组的 items 一律跳过。
func parseMenuItems(raw any) []MenuItem {
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	var out []MenuItem
	for _, entry := range list {
		m, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		item := MenuItem{
			Label:  asString(m["label"]),
			Event:  asString(m["event"]),
			Action: asString(m["action"]),
		}
		if sep, ok := m["separator"].(bool); ok && sep {
			item.Separator = true
		}
		if payload, ok := m["payload"].(map[string]any); ok {
			item.Payload = payload
		}
		item.Items = parseMenuItems(m["items"])
		out = append(out, item)
	}
	return out
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}
