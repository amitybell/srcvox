package main

import (
	"os"
	"strconv"
)

var (
	Env = Environment{
		Minimized:  os.Getenv("Minimized") == "1",
		Demo:       os.Getenv("Demo") == "1",
		InitTab:    os.Getenv("InitTab"),
		InitSbText: os.Getenv("InitSbText"),
		TnetPort:   atoi(os.Getenv("TnetPort")),
	}
)

type Environment struct {
	Minimized  bool   `json:"minimized"`
	Demo       bool   `json:"demo"`
	InitTab    string `json:"initTab"`
	InitSbText string `json:"initSbText"`
	TnetPort   int    `json:"tnetPort"`
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
