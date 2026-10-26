package gofastfyne

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

// loadTrayIcon 从应用根目录下的相对路径加载托盘图标。
// iconPath 为空时返回 nil, nil（使用 Fyne 默认图标）。
func loadTrayIcon(basePath, iconPath string) (fyne.Resource, error) {
	if iconPath == "" {
		return nil, nil
	}
	data, err := os.ReadFile(filepath.Join(basePath, iconPath))
	if err != nil {
		return nil, fmt.Errorf("[GoFast] 加载托盘图标 %s 失败: %w", iconPath, err)
	}
	return fyne.NewStaticResource(filepath.Base(iconPath), data), nil
}
