//go:build windows

package main

import (
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

var (
	steamSearchDirs = func() []string {
		dirs := []string{}
		regPath := `SOFTWARE\WOW6432Node\Valve\Steam`
		reg, err := registry.OpenKey(registry.LOCAL_MACHINE, regPath, registry.QUERY_VALUE)
		if err == nil {
			defer reg.Close()

			dir, _, err := reg.GetStringValue("InstallPath")
			if err == nil {
				dirs = append(dirs, dir)
			}
		}

		defPath := `C:\Program Files (x86)\Steam`
		if len(dirs) != 0 && dirs[0] != defPath {
			dirs = append(dirs, defPath)
		}

		if wd, err := os.Getwd(); err == nil && !strings.HasPrefix(wd, "C:") {
			s := strings.Split(wd, `\`)
			dir := s[0] + `\Program Files (x86)\Steam`
			if dir != defPath {
				dirs = append(dirs, dir)
			}
		}

		return dirs
	}()
)
