package main

import (
	"os"
)

var (
	Env = Environment{
		StartMinimized: os.Getenv("StartMinimized") == "1",
		FakeData:       os.Getenv("FakeData") == "1",
		DefaultTab:     os.Getenv("DefaultTable"),
	}
)

type Environment struct {
	StartMinimized bool   `json:"startMinimized"`
	FakeData       bool   `json:"fakeData"`
	DefaultTab     string `json:"defaultTab"`
}
