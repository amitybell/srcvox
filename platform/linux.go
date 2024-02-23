//go:build linux

package platform

import (
	"os"
	"syscall"
)

const (
	IsLinux   = true
	IsWindows = false
)

var (
	SysProcAttr *syscall.SysProcAttr = nil

	TermSignals = []os.Signal{os.Interrupt, syscall.SIGINT, syscall.SIGTERM}
)
