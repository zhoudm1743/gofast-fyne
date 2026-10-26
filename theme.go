package gofastfyne

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// 主题配置取值。
const (
	ThemeSystem = "system" // 跟随系统（默认）
	ThemeDark   = "dark"
	ThemeLight  = "light"
)

// normalizeTheme 非法取值回退 system。
func normalizeTheme(v string) string {
	switch v {
	case ThemeDark:
		return ThemeDark
	case ThemeLight:
		return ThemeLight
	default:
		return ThemeSystem
	}
}

// forcedTheme 包装基础主题并强制使用指定 variant。
// Fyne 内置主题默认跟随系统偏好，没有强制明暗的开关，需要包一层。
type forcedTheme struct {
	base    fyne.Theme
	variant fyne.ThemeVariant
}

func (t forcedTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.base.Color(name, t.variant)
}

func (t forcedTheme) Font(style fyne.TextStyle) fyne.Resource { return t.base.Font(style) }

func (t forcedTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return t.base.Icon(name) }

func (t forcedTheme) Size(name fyne.ThemeSizeName) float32 { return t.base.Size(name) }

// applyTheme 按配置应用主题；system 不改动当前主题（跟随系统）。
func applyTheme(fapp fyne.App, choice string) {
	switch choice {
	case ThemeDark:
		fapp.Settings().SetTheme(forcedTheme{base: theme.DefaultTheme(), variant: theme.VariantDark})
	case ThemeLight:
		fapp.Settings().SetTheme(forcedTheme{base: theme.DefaultTheme(), variant: theme.VariantLight})
	}
}
