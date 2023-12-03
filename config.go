package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var (
	DefaultConfig = Config{
		AudioDelay:    500 * time.Millisecond,
		AudioLimit:    10 * time.Second,
		AudioLimitTTS: 3 * time.Second,
		TextLimit:     64,
		TnetPort:      31173,
		FirstVoice:    "jenny",
	}
)

type Config struct {
	TnetPort         int             `json:"tnetPort"`
	AudioDelay       time.Duration   `json:"audioDelay"`
	AudioLimit       time.Duration   `json:"audioLimit"`
	AudioLimitTTS    time.Duration   `json:"audioLimitTTS"`
	TextLimit        int             `json:"textLimit"`
	IncludeUsernames map[string]bool `json:"includeUsernames"`
	ExcludeUsernames map[string]bool `json:"excludeUsernames"`
	Hosts            map[string]bool `json:"hosts"`
	FirstVoice       string          `json:"firstVoice"`
	LogLevel         string          `json:"logLevel"`
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
	if p.AudioLimitTTS > 0 {
		c.AudioLimitTTS = p.AudioLimitTTS
	}
	if p.TextLimit > 0 {
		c.TextLimit = p.TextLimit
	}
	if p.IncludeUsernames != nil {
		c.IncludeUsernames = p.IncludeUsernames
	}
	if p.ExcludeUsernames != nil {
		c.ExcludeUsernames = p.ExcludeUsernames
	}
	if p.Hosts != nil {
		c.Hosts = p.Hosts
	}
	if p.LogLevel != "" {
		c.LogLevel = p.LogLevel
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
