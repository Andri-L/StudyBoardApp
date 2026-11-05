package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if isMCPInvocation() {
		if err := runMCPServer(); err != nil {
			fmt.Fprintln(os.Stderr, "StudyBoard MCP:", err)
			os.Exit(1)
		}
		return
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:       "StudyBoard App",
		Width:       1480,
		Height:      940,
		MinWidth:    1080,
		MinHeight:   680,
		AlwaysOnTop: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 243, G: 243, B: 243, A: 1},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "studyboard-app",
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
