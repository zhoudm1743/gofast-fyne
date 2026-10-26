package gofastfyne

import (
	"testing"
)

func TestBuildMainMenu_Structure(t *testing.T) {
	items := []MenuItem{
		{
			Label: "文件",
			Items: []MenuItem{
				{Label: "打开", Event: "file.open"},
				{Separator: true},
				{Label: "退出", Action: ActionQuit},
			},
		},
		{
			Label: "帮助",
			Items: []MenuItem{
				{Label: "关于", Event: "app.about"},
			},
		},
	}

	main := buildMainMenu(items, func(MenuItem) {})

	if len(main.Items) != 2 {
		t.Fatalf("应有 2 个顶层菜单，实际 %d", len(main.Items))
	}
	if main.Items[0].Label != "文件" || main.Items[1].Label != "帮助" {
		t.Errorf("顶层菜单标签错误: %q %q", main.Items[0].Label, main.Items[1].Label)
	}

	file := main.Items[0]
	if len(file.Items) != 3 {
		t.Fatalf("文件子菜单应有 3 项，实际 %d", len(file.Items))
	}
	if !file.Items[1].IsSeparator {
		t.Error("第二项应为分隔符")
	}
	if file.Items[2].Label != "退出" {
		t.Errorf("第三项标签错误: %q", file.Items[2].Label)
	}
}

func TestBuildMenu_ClickFiresCallback(t *testing.T) {
	items := []MenuItem{
		{Label: "打开", Event: "file.open", Payload: map[string]any{"target": "recent"}},
		{Label: "显示", Action: ActionShow},
	}

	var fired []MenuItem
	menu := buildMenu(items, func(mi MenuItem) { fired = append(fired, mi) })

	menu.Items[0].Action()
	menu.Items[1].Action()

	if len(fired) != 2 {
		t.Fatalf("应触发 2 次回调，实际 %d", len(fired))
	}
	if fired[0].Event != "file.open" || fired[0].Payload["target"] != "recent" {
		t.Errorf("首项回调参数错误: %+v", fired[0])
	}
	if fired[1].Action != ActionShow {
		t.Errorf("次项回调参数错误: %+v", fired[1])
	}
}

func TestBuildMenu_EmptyItems(t *testing.T) {
	if m := buildMenu(nil, func(MenuItem) {}); m == nil || len(m.Items) != 0 {
		t.Errorf("空配置应产生空菜单: %+v", m)
	}
	if mm := buildMainMenu(nil, func(MenuItem) {}); mm == nil || len(mm.Items) != 0 {
		t.Errorf("空配置应产生空主菜单: %+v", mm)
	}
}
