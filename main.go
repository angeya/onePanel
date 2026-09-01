package main

import (
	"context"
	"embed"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := InitLogger(); err != nil {
		println("日志初始化失败:", err.Error())
		return
	}
	defer CloseLogger()

	// 单实例检测：已有实例运行时，通知其激活窗口后直接退出
	if !TryAcquireSingleInstance() {
		if err := NotifyExistingInstance(); err != nil {
			LogWarn("通知已有实例激活失败: %v", err)
		} else {
			LogInfo("检测到已有实例，已通知激活其窗口")
		}
		return
	}

	LogInfo("oneWin 应用启动")

	database, err := InitDatabase()
	if err != nil {
		LogError("数据库初始化失败: %v", err)
		return
	}

	app := NewApp(database)
	ptyService := NewPtyService()
	staticServer := NewStaticServer()
	shortcutService := NewShortcutService(database)
	appService := NewAppService(database, staticServer)
	shortcutCmdService := NewShortcutCmdService(database)
	serverListService := NewServerListService(database)
	toolService := NewToolService()
	jsonTreeService := NewJsonTreeService(database)
	settingService := NewSettingService(database)

	var tray *TrayManager
	var hotkey *HotkeyManager

	hotkeyConfig, _ := settingService.GetGlobalHotkey()

	err = wails.Run(&options.App{
		Title:                    "oneWin",
		Width:                    1280,
		Height:                   800,
		Frameless:                true,
		DisableResize:            false,
		EnableDefaultContextMenu: true,
		MinWidth:                 960,
		MinHeight:                640,
		WindowStartState:         options.Normal,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			ptyService.SetContext(ctx)

			// 监听第二次启动的通知，激活主窗口
			StartActivateWatcher(func() {
				runtime.WindowUnminimise(ctx)
				runtime.WindowShow(ctx)
				// Windows 下 ShowWindow 不抢占前台，临时置顶再还原以聚焦窗口
				runtime.WindowSetAlwaysOnTop(ctx, true)
				time.AfterFunc(150*time.Millisecond, func() {
					runtime.WindowSetAlwaysOnTop(ctx, false)
				})
			})

			tray = NewTrayManager(func() {
				runtime.WindowShow(app.ctx)
			}, func() {
				app.QuitApp()
			})
			tray.Start()

			hotkey = NewHotkeyManager(hotkeyConfig, func() {
				runtime.WindowShow(app.ctx)
			})
			if err := hotkey.Start(); err != nil {
				LogWarn("注册全局快捷键失败: %v", err)
			}
		},
		OnBeforeClose: func(ctx context.Context) bool {
			if app.forceQuit {
				return false
			}
			closeAction, _ := app.GetCloseAction()
			if closeAction == "" || closeAction == "ask" {
				runtime.EventsEmit(ctx, "close-requested")
				return true
			}
			if closeAction == "tray" {
				runtime.WindowHide(ctx)
				return true
			}
			return false
		},
		OnShutdown: func(ctx context.Context) {
			if hotkey != nil {
				hotkey.Stop()
			}
			if tray != nil {
				tray.Stop()
			}
			ptyService.StopAll()
			staticServer.Stop()
			database.Close()
		},
		Bind: []interface{}{
			app,
			ptyService,
			shortcutService,
			staticServer,
			appService,
			shortcutCmdService,
			serverListService,
			toolService,
			jsonTreeService,
			settingService,
		},
	})

	if err != nil {
		LogError("应用运行错误: %v", err)
	}
}
