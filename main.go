package main

import (
	"flag"
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
