package gofastfyne

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

// loadResource 从应用根目录下的相对路径加载资源文件（图标等）。
// rel 为空时返回 nil, nil（由调用方决定缺省行为）。
func loadResource(basePath, rel string) (fyne.Resource, error) {
	if rel == "" {
		return nil, nil
	}
	data, err := os.ReadFile(filepath.Join(basePath, rel))
	if err != nil {
		return nil, fmt.Errorf("[GoFast] 加载资源 %s 失败: %w", rel, err)
	}
	return fyne.NewStaticResource(filepath.Base(rel), data), nil
}
