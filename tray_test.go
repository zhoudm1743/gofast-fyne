package gofastfyne

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTrayIcon_OK(t *testing.T) {
	dir := t.TempDir()
	png := []byte{0x89, 0x50, 0x4E, 0x47} // 最小 PNG 头，内容不作为图片解码
	if err := os.WriteFile(filepath.Join(dir, "icon.png"), png, 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := loadTrayIcon(dir, "icon.png")
	if err != nil {
		t.Fatalf("加载托盘图标失败: %v", err)
	}
	if res == nil || res.Name() == "" {
		t.Errorf("资源应非空且有名称: %+v", res)
	}
	if got := res.Content(); len(got) != len(png) {
		t.Errorf("资源内容长度 = %d, 期望 %d", len(got), len(png))
	}
}

func TestLoadTrayIcon_MissingFile(t *testing.T) {
	if _, err := loadTrayIcon(t.TempDir(), "nope.png"); err == nil {
		t.Error("文件不存在应返回错误")
	}
}

func TestLoadTrayIcon_EmptyPath(t *testing.T) {
	res, err := loadTrayIcon(t.TempDir(), "")
	if err != nil || res != nil {
		t.Errorf("空路径应返回 nil, nil，实际 %v, %v", res, err)
	}
}
