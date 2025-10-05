package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"runtime"
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
	if runtime.GOOS == "linux" {
		_ = os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to determine working directory: %v", err)
	}
	log.Printf("starting application in %s", wd)

	if _, err := os.Stat("./tmp"); err != nil {
		if os.IsNotExist(err) {
			log.Print("tmp directory missing, attempting creation")
			if err := os.Mkdir("tmp", os.ModePerm); err != nil {
				log.Fatalf("failed to create tmp directory: %v", err)
			}
			log.Print("tmp directory created successfully")
		}
	} else {
		log.Print("tmp directory already present")
	}
	files, err := filepath.Glob(utils.PathFileTemp("*"))
	if err != nil {
		log.Fatalf("failed to discover tmp files: %v", err)
	}
	if len(files) > 0 {
		log.Printf("removing %d stale tmp file(s)", len(files))
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			log.Fatalf("failed to remove tmp file %s: %v", f, err)
		}
		log.Printf("removed tmp file %s", f)
	}
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	config := &options.App{
		Title:  "wails-go-dekstop-code-interactive",
		Width:  1024,
		Height: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 5},
		OnStartup:        app.startup,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "e3984e08-28dc-4e3d-b70a-45e961589cdc",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			ProgramName:      "wails-go-desktop-code-interactive",
			WebviewGpuPolicy: linux.WebviewGpuPolicyNever,
		},
		Debug: options.Debug{
			OpenInspectorOnStartup: true,
		},
		Windows: &windows.Options{},
	}

	log.Printf("launching Wails with title=%s size=%dx%d", config.Title, config.Width, config.Height)
	log.Printf("single instance lock id=%s", config.SingleInstanceLock.UniqueId)

	err = wails.Run(config)

	if err != nil {
		log.Printf("application exited with error: %v", err)
		println("Error:", err.Error())
		return
	}

	log.Print("application shut down gracefully")
}
