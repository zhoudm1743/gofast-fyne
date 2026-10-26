package gofastfyne

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zhoudm1743/go-fast-framework/config"
	"github.com/zhoudm1743/go-fast-framework/contracts"
)

// newTestConfig 用临时 YAML 文件创建真实的框架配置服务。
func newTestConfig(t *testing.T, yaml string) contracts.Config {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatalf("写入临时配置文件失败: %v", err)
	}
	cfg, err := config.NewConfig(path)
	if err != nil {
		t.Fatalf("创建配置服务失败: %v", err)
	}
	return cfg
}

func TestLoadConfig_Full(t *testing.T) {
	cfg := newTestConfig(t, `
fyne:
  app_id: com.example.app
  window:
    title: 测试窗口
    width: 800
    height: 600
    fixed: true
    close: hide
  menu:
    - label: 文件
      items:
        - label: 打开
          event: file.open
          payload:
            target: recent
        - separator: true
        - label: 退出
          action: quit
    - label: 帮助
      items:
        - label: 关于
          event: app.about
  tray:
    enabled: true
    icon: resources/icon.png
    tooltip: 我的应用
    menu:
      - label: 显示主窗口
        action: show
      - label: 退出
        action: quit
`)

	c := loadConfig(cfg)

	if c.AppID != "com.example.app" {
		t.Errorf("AppID = %q, 期望 com.example.app", c.AppID)
	}
	if c.Window.Title != "测试窗口" || c.Window.Width != 800 || c.Window.Height != 600 {
		t.Errorf("Window = %+v, 期望 测试窗口 800x600", c.Window)
	}
	if !c.Window.Fixed {
		t.Error("Window.Fixed 应为 true")
	}
	if c.Window.Close != CloseHide {
		t.Errorf("Window.Close = %q, 期望 %q", c.Window.Close, CloseHide)
	}

	if len(c.Menu) != 2 {
		t.Fatalf("Menu 应有 2 个顶层菜单，实际 %d", len(c.Menu))
	}
	file := c.Menu[0]
	if file.Label != "文件" || len(file.Items) != 3 {
		t.Fatalf("文件菜单解析错误: %+v", file)
	}
	open := file.Items[0]
	if open.Label != "打开" || open.Event != "file.open" {
		t.Errorf("打开菜单项解析错误: %+v", open)
	}
	if open.Payload["target"] != "recent" {
		t.Errorf("Payload 解析错误: %+v", open.Payload)
	}
	if !file.Items[1].Separator {
		t.Error("第二项应为分隔符")
	}
	quit := file.Items[2]
	if quit.Label != "退出" || quit.Action != ActionQuit {
		t.Errorf("退出菜单项解析错误: %+v", quit)
	}
	if c.Menu[1].Label != "帮助" || c.Menu[1].Items[0].Event != "app.about" {
		t.Errorf("帮助菜单解析错误: %+v", c.Menu[1])
	}

	if !c.Tray.Enabled {
		t.Error("Tray.Enabled 应为 true")
	}
	if c.Tray.Icon != "resources/icon.png" || c.Tray.Tooltip != "我的应用" {
		t.Errorf("Tray 解析错误: %+v", c.Tray)
	}
	if len(c.Tray.Menu) != 2 || c.Tray.Menu[0].Action != ActionShow {
		t.Errorf("托盘菜单解析错误: %+v", c.Tray.Menu)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	c := loadConfig(newTestConfig(t, ""))

	if c.AppID == "" {
		t.Error("AppID 应有默认值")
	}
	if c.Window.Title == "" || c.Window.Width <= 0 || c.Window.Height <= 0 {
		t.Errorf("窗口应有默认标题与尺寸: %+v", c.Window)
	}
	if c.Window.Close != CloseQuit {
		t.Errorf("Window.Close 默认应为 %q，实际 %q", CloseQuit, c.Window.Close)
	}
	if c.Tray.Enabled {
		t.Error("托盘默认应关闭")
	}
	if len(c.Menu) != 0 {
		t.Errorf("默认无菜单，实际 %d 项", len(c.Menu))
	}
}

func TestLoadConfig_InvalidCloseFallsBack(t *testing.T) {
	c := loadConfig(newTestConfig(t, "fyne:\n  window:\n    close: minimize\n"))
	if c.Window.Close != CloseQuit {
		t.Errorf("非法 close 值应回退 %q，实际 %q", CloseQuit, c.Window.Close)
	}
}

func TestLoadConfig_MalformedMenuSkipped(t *testing.T) {
	cfg := newTestConfig(t, `
fyne:
  menu:
    - 123
    - label: 有效
      items: "不是数组"
    - label: 正常
      items:
        - label: 子项
          action: toggle
`)
	c := loadConfig(cfg)
	if len(c.Menu) != 2 {
		t.Fatalf("畸形项应被跳过，期望 2 个菜单，实际 %d: %+v", len(c.Menu), c.Menu)
	}
	if len(c.Menu[0].Items) != 0 {
		t.Errorf("非数组 items 应解析为空，实际 %+v", c.Menu[0].Items)
	}
	if c.Menu[1].Items[0].Action != ActionToggle {
		t.Errorf("子项解析错误: %+v", c.Menu[1].Items[0])
	}
}
