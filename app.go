//go:build windows

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
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
	ctx context.Context
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
	Txt      string `json:"txt"`
	Stdout   string `json:"out"`
	Stderr   string `json:"errout"`
	Language string `json:"lang"`
	Tipe     string `json:"type"`
}

func (a *App) CheckFileExecutable(name []string) (all []string) {
	for _, v := range name {
		cmd := exec.Command("where", v)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: 0x08000000,
		}
		err := cmd.Run()
		if err == nil {
			all = append(all, v)
		}
	}
	return all
}

func (a *App) RunFileExecutable(data Data) (*Data, error) {
	var args string = "-"
	data.Txt = strings.TrimSpace(data.Txt)
	data.Language = strings.TrimSpace(data.Language)
	data.Tipe = strings.TrimSpace(data.Tipe)
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

	out, errout, err := utils.Shellout(data.Language, utils.PathFileTemp(filename))
	if data.Language == "go" {
		out, errout, err = utils.Shellout(data.Language, args, utils.PathFileTemp(filename))
	}

	if err != nil {
		log.Printf("error shell: %v\n", err)
	}

	fmt.Println("--- stdout ---")
	fmt.Println(out)
	fmt.Println("--- stderr ---")
	fmt.Println(errout)
	data.Stderr = errout
	data.Stdout = out
	return &data, nil
}
