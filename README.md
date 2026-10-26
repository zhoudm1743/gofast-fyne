# gofast-fyne

[GoFast](https://github.com/zhoudm1743/go-fast-framework) 桌面 GUI 驱动插件，基于 [Fyne](https://fyne.io/)（纯 Go、跨平台、无 Webview）。

定位对标 `gofast-gin`：插件自带 ServiceProvider，不改框架一行代码，通过 `app.MustMake("gui")` 获取桌面 GUI 服务，框架的配置、日志、数据库、事件总线等全部能力可直接在 UI 代码中使用。

## 特性

- **生命周期对接**：`Run()` 阻塞运行 Fyne 事件循环（主线程）；窗口关闭 → 框架 `Shutdown()`（幂等），框架 `Shutdown()` → Fyne `Quit()`
- **延迟加载**：实现 `foundation.DeferredProvider`，仅当首次解析 `"gui"` 服务时初始化，纯 CLI / 队列 worker 场景零 GUI 开销
- **配置驱动菜单/托盘**：`config.yaml` 声明主窗口菜单与系统托盘，支持子菜单、分隔符、内置动作（show/hide/toggle/quit）与自定义事件
- **事件总线桥接**：菜单/托盘点击 `Emit` 字符串事件进框架事件总线；`On` 监听字符串事件、`Listen` 监听类型化框架事件，回调保证切回 Fyne UI 线程执行
- **窗口级事件**：`window.shown` / `window.hidden` / `window.closed` 经框架事件总线派发，`On` 即可订阅（回调在 UI 线程）
- **WindowController 契约**：`Title()` + `View(g Gui)` 两个方法即是一个窗口页面控制器，由应用侧装配进导航（如 AppTabs），对照 HTTP 的 `contracts.Controller`

## 安装

```bash
go get github.com/zhoudm1743/gofast-fyne
```

> Linux 编译需要 OpenGL/X11 开发库（`libgl1-mesa-dev xorg-dev` 等）。

## 配置（config/config.yaml）

```yaml
fyne:
  app_id: com.example.desktop      # 应用 ID（影响偏好存储位置）
  window:
    title: 我的应用
    width: 1024
    height: 768
    fixed: false                   # 固定尺寸
    close: hide                    # quit（默认，关闭即退出）| hide（关闭仅隐藏，配合托盘）
  menu:
    - label: 文件
      items:
        - label: 打开
          event: file.open         # 点击 → 向框架事件总线派发 "file.open"
          payload:                 # 可选载荷
            target: recent
        - separator: true
        - label: 退出
          action: quit             # 内置动作：show / hide / toggle / quit
  tray:
    enabled: true
    icon: resources/icon.png       # 相对应用根目录；可省略
    menu:
      - label: 显示主窗口
        action: show
      - label: 退出
        action: quit
```

## 用法

bootstrap 注册 Provider（放在 `gohttp.ServiceProvider` 等之后即可）：

```go
import gofastfyne "github.com/zhoudm1743/gofast-fyne"

app.SetProviders([]foundation.ServiceProvider{
    // ... 其他 Provider
    &gofastfyne.ServiceProvider{}, // 桌面 GUI 驱动（延迟加载）
})
```

main 运行（必须在主线程，且先锁定 OS 线程——bootstrap 的后台 goroutine 可能把主 goroutine 挤离主线程，导致 GLFW 初始化失败）：

```go
func main() {
    // Fyne/GLFW 要求 UI 操作发生在主线程，先锁定
    runtime.LockOSThread()

    app := bootstrap.Boot()

    gui := app.MustMake("gui").(gofastfyne.Gui)

    gui.On("file.open", func(p map[string]any) {
        // 回调已在 UI 线程，可直接操作界面
    })
    gui.SetContent(widget.NewLabel("Hello GoFast"))
    gui.Listen(&events.OrderShipped{}, func(args ...any) { /* 框架事件 → UI */ })

    _ = gui.Run() // 阻塞，窗口关闭后自动触发框架 Shutdown
}
```

## API 摘要

| 方法 | 说明 |
|------|------|
| `Run()` | 显示主窗口并阻塞运行事件循环，退出后幂等触发框架 Shutdown |
| `Quit()` | 退出 Fyne 应用（幂等） |
| `Window()` / `NewWindow(title)` | 主窗口 / 新建窗口 |
| `SetContent(c)` | 设置主窗口内容 |
| `ShowMainWindow()` / `HideMainWindow()` / `ToggleMainWindow()` / `Visible()` | 主窗口显隐控制 |
| `Emit(name, payload)` | 向框架事件总线派发字符串命名的 UI 事件 |
| `On(name, fn)` | 监听字符串事件，回调在 UI 线程执行 |
| `Listen(eventer, fn)` | 监听类型化框架事件，回调在 UI 线程执行 |

## 说明

- 托盘依赖桌面环境的系统托盘支持；非桌面驱动（含测试驱动）下自动跳过，不报错
- `fyne.window.close: hide` 时建议同时启用托盘，否则窗口隐藏后无法唤回
- 交叉编译 / 打包参考 [fyne-cross](https://github.com/fyne-io/fyne-cross)
