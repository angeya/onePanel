package main

import (
	"context"
	"embed"

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
	settingService := NewSettingService(database)

	var tray *TrayManager
	var hotkey *HotkeyManager

	err = wails.Run(&options.App{
		Title:            "oneWin",
		Width:            1280,
		Height:           800,
		Frameless:        true,
		DisableResize:    false,
		MinWidth:         960,
		MinHeight:        640,
		WindowStartState: options.Normal,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			ptyService.SetContext(ctx)

			if err := InitContextMenuControl(ctx); err != nil {
				LogWarn("初始化上下文菜单控制失败: %v", err)
			} else {
				allowDebugVal, _ := database.GetConfig("allow_debug")
				if allowDebugVal == "true" {
					if err := SetContextMenuEnabled(ctx, true); err != nil {
						LogWarn("启用上下文菜单失败: %v", err)
					}
				} else {
					if err := SetContextMenuEnabled(ctx, false); err != nil {
						LogWarn("禁用上下文菜单失败: %v", err)
					}
				}
			}

			tray = NewTrayManager(func() {
				runtime.WindowShow(app.ctx)
			}, func() {
				app.QuitApp()
			})
			tray.Start()

			hotkey = NewHotkeyManager(func() {
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
			CleanupContextMenuControl()
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
			settingService,
		},
	})

	if err != nil {
		LogError("应用运行错误: %v", err)
	}
}
