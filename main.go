package main

import (
	"flag"
	"os"
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
	if logw.F != nil {
		os.Stderr = logw.F
	}

	app := NewApp(paths)
	defer app.Close()

	if err := app.Run(); err != nil {
		Logs.Println("Error:", err.Error())
	}
}
