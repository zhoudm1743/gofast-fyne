package gofastfyne

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestNormalizeTheme(t *testing.T) {
	cases := map[string]string{
		"dark":   ThemeDark,
		"light":  ThemeLight,
		"system": ThemeSystem,
		"":       ThemeSystem,
		"weird":  ThemeSystem,
	}
	for in, want := range cases {
		if got := normalizeTheme(in); got != want {
			t.Errorf("normalizeTheme(%q) = %q, 期望 %q", in, got, want)
		}
	}
}

func TestApplyTheme_ForcesVariant(t *testing.T) {
	a := test.NewApp()
	bg := theme.ColorNameBackground

	// 强制 light：无论询问哪个 variant，都应返回 light 配色
	applyTheme(a, ThemeLight)
	got := a.Settings().Theme().Color(bg, theme.VariantDark)
	want := theme.DefaultTheme().Color(bg, theme.VariantLight)
	if got != want {
		t.Errorf("强制 light 后 Color(bg, VariantDark) = %v, 期望 light 配色 %v", got, want)
	}

	// 强制 dark
	applyTheme(a, ThemeDark)
	got = a.Settings().Theme().Color(bg, theme.VariantLight)
	want = theme.DefaultTheme().Color(bg, theme.VariantDark)
	if got != want {
		t.Errorf("强制 dark 后 Color(bg, VariantLight) = %v, 期望 dark 配色 %v", got, want)
	}
}

func TestApplyTheme_SystemLeavesDefault(t *testing.T) {
	a := test.NewApp()
	before := a.Settings().Theme()
	applyTheme(a, ThemeSystem)
	if a.Settings().Theme() != before {
		t.Error("system 主题不应修改当前主题")
	}
}
