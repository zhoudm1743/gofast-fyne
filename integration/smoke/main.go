//go:build integration

// smoke 是真实 GLFW 驱动的冒烟程序：以进程主线程跑完整 UI 流程。
// 由集成测试通过 `go run -tags integration ./integration/smoke` 启动，
// 不要直接运行。
package main

import (
	"fmt"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

func main() {
	gui, err := gofastfyne.New(app.NewWithID("com.gofast.smoke"), gofastfyne.Config{
		AppID:  "com.gofast.smoke",
		Theme:  gofastfyne.ThemeSystem,
		Window: gofastfyne.WindowConfig{Title: "smoke", Width: 400, Height: 300, Close: gofastfyne.CloseQuit},
	}, ".", nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "new gui:", err)
		os.Exit(1)
	}

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
	fmt.Println("SMOKE-OK")
}
