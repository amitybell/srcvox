package main

import (
	"flag"

	"github.com/amitybell/srcvox/config"
	"github.com/amitybell/srcvox/logs"
)

var (
	paths = config.DefaultPaths
	Logs  = logs.AppLogger()
)

func main() {
	defer Logs.Close()

	flag.Parse()

	app := NewApp(paths)
	defer app.Close()

	if err := app.Run(); err != nil {
		Logs.Println("Error:", err.Error())
	}
}
