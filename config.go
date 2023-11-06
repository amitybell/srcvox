package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var (
	DefaultConfig = Config{
		AudioDelay: 500 * time.Millisecond,
		AudioLimit: 10 * time.Second,
		TnetPort:   31173,
	}
)

type Config struct {
	TnetPort         int             `json:"tnetPort"`
	AudioDelay       time.Duration   `json:"audioDelay"`
	AudioLimit       time.Duration   `json:"audioLimit"`
	IncludeUsernames map[string]bool `json:"includeUsernames"`
	ExcludeUsernames map[string]bool `json:"excludeUsernames"`
}

func (c Config) Merge(p Config) Config {
	if p.TnetPort > 0 {
		c.TnetPort = p.TnetPort
	}
	if p.AudioDelay > 0 {
		c.AudioDelay = p.AudioDelay
	}
	if p.AudioLimit > 0 {
		c.AudioLimit = p.AudioLimit
	}
	if p.IncludeUsernames != nil {
		c.IncludeUsernames = p.IncludeUsernames
	}
	if p.ExcludeUsernames != nil {
		c.ExcludeUsernames = p.ExcludeUsernames
	}
	return c
}

func readConfig(fn string) (Config, error) {
	cfg := DefaultConfig
	f, err := os.Open(fn)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("readConfig: open: %w", err)
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("readConfig: parse: %w", err)
	}
	return cfg, nil
}
