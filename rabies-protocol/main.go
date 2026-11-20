package main

import (
	"embed"
	"rabies-protocol/internal/wails"

	wailsruntime "github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := wails.NewApp()

	err := wailsruntime.Run(&options.App{
		Title:  "RabiesProtocol — Протокол осмотра трупа",
		Width:  1024,
		Height: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.Startup,
		OnShutdown: app.Shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		panic(err)
	}
}