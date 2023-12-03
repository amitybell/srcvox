//go:build linux

package main

import (
	"log/slog"
	"os"
	"path/filepath"
)

var (
	steamSearchDirs = func() []string {
		home, err := os.UserHomeDir()
		if err != nil {
			Logs.Println("Cannot get user home dir:", err)
			return nil
		}
		dirs := []string{
			filepath.Join(home, ".local/share/Steam"),
			filepath.Join(home, ".var/app/com.valvesoftware.Steam/.local/share/Steam"),
		}
		Logs.Debug("steamSearchDirs", slog.Any("dirs", dirs))
		return dirs
	}()
)
