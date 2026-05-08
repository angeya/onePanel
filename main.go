package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	ptyService := NewPtyService()
	shortcutService := NewShortcutService()
	historyService := NewHistoryService()
	staticServer := NewStaticServer()
	appService := NewAppService()
	shortcutCmdService := NewShortcutCmdService()
	toolService := NewToolService()
	settingService := NewSettingService()

	err := wails.Run(&options.App{
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
			if err := InitDatabase(); err != nil {
				println("数据库初始化失败:", err.Error())
			}
		},
		OnShutdown: func(ctx context.Context) {
			ptyService.StopAll()
			staticServer.Stop()
			CloseDatabase()
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
		println("Error:", err.Error())
	}
}
