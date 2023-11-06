package main

import (
	"flag"
	"github.com/amitybell/srcvox/files"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

func main() {
	flag.Parse()

	paths, err := NewPaths("", "")
	if err != nil {
		Logs.Fatal(err)
	}

	logw := NewLogWriter(paths.LogsFn)
	defer logw.Close()
	Logs.SetOutput(logw)

	app := NewApp(paths)
	defer app.Close()

	err = wails.Run(&options.App{
		Title:            "SrcVox",
		Width:            800,
		Height:           600,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		Linux: &linux.Options{
			WebviewGpuPolicy: linux.WebviewGpuPolicyNever,
			Icon:             files.EmblemPNG,
		},
		Windows: &windows.Options{
			WebviewUserDataPath:  paths.WebviewDataDir,
			WebviewGpuIsDisabled: true,
		},
		OnStartup: app.OnStartup,
		Bind:      []any{app.API},
		ErrorFormatter: func(err error) any {
			switch err := err.(type) {
			case nil:
				return nil
			case *AppError:
				return err
			default:
				return &AppError{Message: err.Error()}
			}
		},
		AssetServer: &assetserver.Options{
			Assets:  assetsFS,
			Handler: app,
		},
		WindowStartState: func() options.WindowStartState {
			if Env.StartMinimized {
				return options.Minimised
			}
			return options.Normal
		}(),
	})

	if err != nil {
		Logs.Println("Error:", err.Error())
	}
}
