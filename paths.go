package main

import (
	"fmt"
	"github.com/adrg/xdg"
	"path/filepath"
)

type Paths struct {
	ConfigDir      string
	ConfigFn       string
	DataDir        string
	WebviewDataDir string
	DBDir          string
	LogsFn         string
}

func NewPaths(configDir, dataDir string) (*Paths, error) {
	if configDir == "" {
		dir, err := xdg.ConfigFile("srcvox")
		if err != nil {
			return nil, fmt.Errorf("Cannot init config dir: %w", err)
		}
		configDir = dir
	}

	if dataDir == "" {
		dir, err := xdg.DataFile("srcvox")
		if err != nil {
			return nil, fmt.Errorf("Cannot init data dir: %w", err)
		}
		dataDir = dir
	}

	return &Paths{
		ConfigDir:      configDir,
		ConfigFn:       filepath.Join(configDir, "config.json"),
		DataDir:        dataDir,
		WebviewDataDir: filepath.Join(dataDir, "webview"),
		DBDir:          filepath.Join(dataDir, "data.pb"),
		LogsFn:         filepath.Join(dataDir, "logs.txt"),
	}, nil
}
