package main

import (
	"os"
)

var (
	Env = Environment{
		Minimized:  os.Getenv("Minimized") == "1",
		Demo:       os.Getenv("Demo") == "1",
		InitTab:    os.Getenv("InitTab"),
		InitSbText: os.Getenv("InitSbText"),
	}
)

type Environment struct {
	Minimized  bool   `json:"minimized"`
	Demo       bool   `json:"demo"`
	InitTab    string `json:"initTab"`
	InitSbText string `json:"initSbText"`
}
