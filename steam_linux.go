//go:build linux

package main

import (
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
		return []string{
			filepath.Join(home, ".local/share/Steam"),
			filepath.Join(home, ".var/app/com.valvesoftware.Steam/.local/share/Steam"),
		}
	}()
)
