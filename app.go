//go:build windows

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"wails-go-desktop-code-interactive/utils"

	"github.com/wailsapp/wails/v2/pkg/options"
	ru "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	wailsContext       *context.Context
	secondInstanceArgs []string
)

// App struct
type App struct {
	ctx         context.Context
	runtimeRoot string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	wailsContext = &ctx
	if err := a.prepareRuntimeBundles(); err != nil {
		log.Printf("failed to prepare bundled runtimes: %v", err)
	}
}

func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	secondInstanceArgs = secondInstanceData.Args

	println("user opened second instance", strings.Join(secondInstanceData.Args, ","))
	println("user opened second from", secondInstanceData.WorkingDirectory)
	ru.WindowUnminimise(*wailsContext)
	ru.Show(*wailsContext)
	go ru.EventsEmit(*wailsContext, "launchArgs", secondInstanceArgs)
}

type Data struct {
	Txt              string `json:"txt"`
	Stdout           string `json:"out"`
	Stderr           string `json:"errout"`
	Language         string `json:"lang"`
	Tipe             string `json:"type"`
	ExecMode         string `json:"execMode"`
	BundledRuntime   string `json:"bundledRuntime"`
	CustomExecutable string `json:"customExecutable"`
	CustomWorkingDir string `json:"customWorkingDir"`
	PreferBundled    bool   `json:"preferBundled"`
}

func (a *App) CheckFileExecutable(name []string) (all []string) {
	availability := a.ListLanguageAvailability(name)
	combined := append(availability.System, availability.Bundled...)
	return dedupeStrings(combined)
}

func (a *App) RunFileExecutable(data Data) (*Data, error) {
	args := "-"
	data.Txt = strings.TrimSpace(data.Txt)
	data.Language = strings.TrimSpace(data.Language)
	data.Tipe = strings.TrimSpace(data.Tipe)
	data.ExecMode = strings.ToLower(strings.TrimSpace(data.ExecMode))
	data.BundledRuntime = strings.TrimSpace(data.BundledRuntime)
	data.CustomExecutable = strings.TrimSpace(data.CustomExecutable)
	data.CustomWorkingDir = strings.TrimSpace(data.CustomWorkingDir)

	switch data.ExecMode {
	case "bundled", "custom":
		// keep as-is
	case "default", "system", "":
		data.ExecMode = "default"
	default:
		data.ExecMode = "default"
	}

	if data.ExecMode == "custom" && data.CustomExecutable == "" {
		return &Data{Stdout: "Nothing", Stderr: "Nothing"}, fmt.Errorf("custom executable path is required when execMode is custom")
	}

	if data.ExecMode == "bundled" {
		if data.BundledRuntime == "" {
			if runtimes := a.ListBundledRuntimes(); len(runtimes) > 0 {
				data.BundledRuntime = runtimes[0]
			} else {
				return &Data{Stdout: "Nothing", Stderr: "Nothing"}, fmt.Errorf("bundled runtime requested but none are available")
			}
		}
		data.PreferBundled = false
	}

	if data.ExecMode == "custom" {
		data.PreferBundled = false
	}
	if data.Language == "" || data.Txt == "" || data.Tipe == "" {
		return &Data{Stdout: "Nothing", Stderr: "Nothing"}, fmt.Errorf("Something went wrong, when data empty")
	}

	if !utils.CheckIsNotData([]string{"php", "node", "go"}, data.Language) {
		return &Data{Stdout: "Nothing", Stderr: "Nothing"}, fmt.Errorf("Something went wrong, check language is empty")
	}

	if !utils.CheckIsNotData([]string{"repl", "stq"}, data.Tipe) {
		return &Data{Stdout: "Nothing", Stderr: "Nothing"}, fmt.Errorf("Something Went Wrong, Check Type is exist")
	}

	filename := "index-" + utils.StringWithCharset(5) + ".js"
	if data.Language == "php" {
		filename = "index-" + utils.StringWithCharset(5) + ".php"
	}
	if data.Language == "go" {
		args = "run"
		filename = "main-" + utils.StringWithCharset(5) + ".go"
	}
	if data.Tipe == "stq" {
		if data.Language == "go" {
			data.Txt = data.Txt + utils.TxtGo
		}
		if data.Language == "php" {
			data.Txt = data.Txt + utils.TxtPHP
		}
		if data.Language == "node" {
			data.Txt = data.Txt + utils.TxtJS
		}
	}

	err := os.WriteFile(filename, []byte(data.Txt), 0755)
	if err != nil {
		log.Print("unable to write file: ", err)
		return &Data{Stdout: "Nothing", Stderr: "Nothing"}, fmt.Errorf("Something went wrong, unable to write file")
	}
	err = utils.MoveFile(filename, utils.PathFileTemp(filename))
	if err != nil {
		log.Print("error movefile: ", err)
		return &Data{Stdout: "Nothing", Stderr: "Nothing"}, fmt.Errorf("Something went wrong, unable move file")
	}

	commandPath, workingDir, resolveErr := a.resolveExecutionTarget(data)
	if resolveErr != nil {
		log.Printf("resolve execution target failed: %v", resolveErr)
		return &Data{Stdout: "Nothing", Stderr: "Unable to resolve executable"}, fmt.Errorf("unable to resolve executable: %w", resolveErr)
	}

	var execCfg *utils.ExecConfig
	if workingDir != "" {
		execCfg = &utils.ExecConfig{Dir: workingDir}
	}

	out, errout, err := utils.Shellout(commandPath, execCfg, utils.PathFileTemp(filename))
	if data.Language == "go" {
		out, errout, err = utils.Shellout(commandPath, execCfg, args, utils.PathFileTemp(filename))
	}

	if err != nil {
		log.Printf("error shell: %v\n", err)
	}
	data.Stderr = errout
	data.Stdout = out
	return &data, nil
}
