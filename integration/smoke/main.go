//go:build integration

// smoke 是真实 GLFW 驱动的冒烟程序：以进程主线程跑完整 UI 流程。
// 由集成测试通过 `go run -tags integration ./integration/smoke <mode>` 启动，
// 不要直接运行。
//
// 模式：
//   - quit：New 直建 Gui → OpenWindow → CloseWindow → Quit → Run 返回
//   - sigterm：走 ServiceProvider 装配（含 OnShutdown 双向挂钩），
//     向自身发送 SIGTERM，复现并验证信号路径：
//     Fyne 优雅退出 → Run 返回 → 框架 Shutdown → Quit 守卫不二次清理 GLFW
package main

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"github.com/zhoudm1743/go-fast-framework/config"
	"github.com/zhoudm1743/go-fast-framework/event"
	"github.com/zhoudm1743/go-fast-framework/foundation"
	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

func main() {
	mode := "quit"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	gui, err := buildGui(mode)
	if err != nil {
		fmt.Fprintln(os.Stderr, "build gui:", err)
		os.Exit(1)
	}

	switch mode {
	case "quit":
		opened := gui.OpenWindow("detail", "detail", widget.NewLabel("sub"))
		go func() {
			time.Sleep(2 * time.Second)
			// 窗口操作属 UI 线程 API，其他线程需经 fyne.Do 派发（Quit 自身已内部派发）
			fyne.Do(func() { gui.CloseWindow("detail") })
			gui.Quit()
		}()
		if err := gui.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "run:", err)
			os.Exit(1)
		}
		if opened == nil {
			os.Exit(1)
		}

	case "sigterm":
		go func() {
			time.Sleep(2 * time.Second)
			// Fyne 驱动自行捕获 SIGTERM 并优雅退出（关窗→drain→GLFW Terminate）
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()
		// Run 返回后会经 shutdownOnce 触发框架 Shutdown → OnShutdown(g.Quit)，
		// 此处即验证该路径不 panic。
		if err := gui.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "run:", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintln(os.Stderr, "unknown mode:", mode)
		os.Exit(2)
	}

	fmt.Println("SMOKE-OK")
}

// buildGui 按模式创建 Gui；sigterm 模式走 ServiceProvider 完整装配
// （含框架 Shutdown 双向挂钩），以覆盖信号驱动的退出链路。
func buildGui(mode string) (gofastfyne.Gui, error) {
	if mode != "sigterm" {
		return gofastfyne.New(app.NewWithID("com.gofast.smoke"), gofastfyne.Config{
			AppID:  "com.gofast.smoke",
			Theme:  gofastfyne.ThemeSystem,
			Window: gofastfyne.WindowConfig{Title: "smoke", Width: 400, Height: 300, Close: gofastfyne.CloseQuit},
		}, ".", nil)
	}

	fapp := foundation.NewApplication(".")
	fapp.SetProviders([]foundation.ServiceProvider{
		&config.ServiceProvider{},
		&event.ServiceProvider{},
		&gofastfyne.ServiceProvider{},
	})
	fapp.Boot()
	return fapp.MustMake("gui").(gofastfyne.Gui), nil
}
