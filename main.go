package main

import (
	"context"
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	database, err := InitDatabase()
	if err != nil {
		println("数据库初始化失败:", err.Error())
		return
	}

	app := NewApp()
	ptyService := NewPtyService()
	staticServer := NewStaticServer()
	shortcutService := NewShortcutService(database)
	historyService := NewHistoryService(database)
	appService := NewAppService(database, staticServer)
	shortcutCmdService := NewShortcutCmdService(database)
	toolService := NewToolService()
	settingService := NewSettingService(database)

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
		},
		OnShutdown: func(ctx context.Context) {
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
