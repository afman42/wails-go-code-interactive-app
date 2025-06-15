package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"wails-go-desktop-code-interactive/utils"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {

	if _, err := os.Stat("./tmp"); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir("tmp", os.ModePerm); err != nil {
				log.Fatal(err)
			}
		}
	}
	files, err := filepath.Glob(utils.PathFileTemp("*"))
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			panic(err)
		}
	}
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "wails-go-dekstop-code-interactive",
		Width:  1024,
		Height: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 5},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			ProgramName: "wails-go-desktop-code-interactive",
		},
		Debug: options.Debug{
			OpenInspectorOnStartup: true,
		},
		Windows: &windows.Options{},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
