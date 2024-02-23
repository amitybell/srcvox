//go:build windows

package platform

import (
	"os"
	"syscall"
)

const (
	IsLinux   = false
	IsWindows = true
)

var (
	SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}

	TermSignals = []os.Signal{os.Interrupt, syscall.SIGINT, syscall.SIGTERM}
)
