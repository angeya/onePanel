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

	err := wails.Run(&options.App{
		Title:  "onePanel",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			ptyService.SetContext(ctx)
		},
		OnShutdown: func(ctx context.Context) {
			ptyService.Stop()
		},
		Bind: []interface{}{
			app,
			ptyService,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
