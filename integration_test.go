//go:build integration

package gofastfyne_test

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestIntegration_RealDriver 真实 GLFW 驱动冒烟。
// Fyne 要求 UI 调用发生在进程主线程，而 go test 的测试函数跑在
// 测试框架 goroutine 中，无法承担该角色——因此真正的 UI 流程放在
// integration/smoke 这个独立 main 程序里，本测试以子进程方式运行它。
//
// 需要显示服务器（X11/Wayland），运行：go test -tags integration ./...
func TestIntegration_RealDriver(t *testing.T) {
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		t.Skip("无显示服务器，跳过真实驱动集成测试")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "-tags", "integration", "./integration/smoke")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	t.Logf("smoke output:\n%s", out)
	if err != nil {
		t.Fatalf("冒烟程序失败: %v", err)
	}
	if !strings.Contains(string(out), "SMOKE-OK") {
		t.Fatal("冒烟程序未输出 SMOKE-OK")
	}
}
