package main

import (
	"context"
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	database, err := InitDatabase()
	if err != nil {
		println("数据库初始化失败:", err.Error())
		return
	}

	app := NewApp(database)
	ptyService := NewPtyService()
	staticServer := NewStaticServer()
	shortcutService := NewShortcutService(database)
	historyService := NewHistoryService(database)
	appService := NewAppService(database, staticServer)
	shortcutCmdService := NewShortcutCmdService(database)
	toolService := NewToolService()
	settingService := NewSettingService(database)

	var tray *TrayManager
	var hotkey *HotkeyManager

	err = wails.Run(&options.App{
		Title:  "oneWin",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			ptyService.SetContext(ctx)

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
				fmt.Println("注册全局快捷键失败:", err)
			}
		},
		OnBeforeClose: func(ctx context.Context) bool {
			if app.forceQuit {
				return false
			}
			closeAction := app.GetCloseAction()
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
			historyService,
			staticServer,
			appService,
			shortcutCmdService,
			toolService,
			settingService,
		},
	})

	if err != nil {
		fmt.Println("Error:", err.Error())
	}
}
