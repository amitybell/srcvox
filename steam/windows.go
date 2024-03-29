//go:build windows

package steam

import (
	"log/slog"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

var (
	steamSearchDirs = func() []string {
		dirs := []string{}
		regPath := `SOFTWARE\WOW6432Node\Valve\Steam`
		reg, err := registry.OpenKey(registry.LOCAL_MACHINE, regPath, registry.QUERY_VALUE)
		if err != nil {
			Logs.Debug("Cannot read registry",
				slog.String("key", regPath),
				slog.Any("error", err))
		} else {
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

		Logs.Debug("steamSearchDirs", slog.Any("dirs", dirs))

		return dirs
	}()
)
