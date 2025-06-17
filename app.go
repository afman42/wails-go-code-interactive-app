package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
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

type Meta struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

type ResponseData struct {
	Meta *Meta `json:"meta"`
	Data *Data `json:"data"`
}

func (a *App) CheckFileExecutable(name []string) (all []string) {
	os := runtime.GOOS
	for _, v := range name {
		if os == "windows" {
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
		if os == "linux" {
			cmd := exec.Command("which", v)
			err := cmd.Run()
			if err == nil {
				all = append(all, v)
			}
		}
	}
	return all
}

func (a *App) RunFileExecutable(data Data) ResponseData {
	var args string = "-"
	data.Txt = strings.TrimSpace(data.Txt)
	data.Language = strings.TrimSpace(data.Language)
	data.Tipe = strings.TrimSpace(data.Tipe)
	if data.Language == "" || data.Txt == "" || data.Tipe == "" {
		data.Stdout = "Nothing"
		data.Stderr = "Nothing"
		return ResponseData{
			Meta: &Meta{
				StatusCode: http.StatusBadRequest,
				Message:    "Something Went Wrong, Check Language is exist",
			},
			Data: &data,
		}

	}
	if !utils.CheckIsNotData([]string{"php", "node", "go"}, data.Language) {
		data.Stdout = "Nothing"
		data.Stderr = "Nothing"
		return ResponseData{
			Meta: &Meta{
				StatusCode: http.StatusBadRequest,
				Message:    "Something Went Wrong, Check Language is exist",
			},
			Data: &data,
		}
	}
	if !utils.CheckIsNotData([]string{"repl", "stq"}, data.Tipe) {
		data.Stdout = "Nothing"
		data.Stderr = "Nothing"
		return ResponseData{
			Meta: &Meta{
				StatusCode: http.StatusBadRequest,
				Message:    "Something Went Wrong, Check Type is exist",
			},
			Data: &data,
		}
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
		data.Stderr = "Nothing"
		data.Stdout = "Nothing"
		return ResponseData{
			Meta: &Meta{
				StatusCode: http.StatusBadRequest,
				Message:    "Something Went Wrong",
			},
			Data: &data,
		}
	}
	err = utils.MoveFile(filename, utils.PathFileTemp(filename))
	if err != nil {
		log.Print("error movefile: ", err)
		data.Stderr = "Nothing"
		data.Stdout = "Nothing"
		return ResponseData{
			Meta: &Meta{
				StatusCode: http.StatusBadRequest,
				Message:    "Something Went Wrong",
			},
			Data: &data,
		}
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
	return ResponseData{
		Meta: &Meta{
			StatusCode: http.StatusOK,
			Message:    "Success",
		},
		Data: &data,
	}
}
