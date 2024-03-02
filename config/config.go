package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

var DefaultPaths = MustNewPaths("", "")

var DefaultConfig = func() Config {
	var cfg Config
	json.NewDecoder(strings.NewReader(os.Getenv("cfg"))).Decode(&cfg)
	def := Config{
		AudioDelay:    Dur{D: 250 * time.Millisecond},
		AudioLimit:    Dur{D: 3 * time.Second},
		AudioLimitTTS: Dur{D: 3 * time.Second},
		TextLimit:     64,
		Netcon: ConnInfo{
			Host: "127.0.0.1",
			Port: 31173,
		},
		TnetPort:         31173,
		FirstVoice:       "jenny",
		RateLimit:        Dur{D: 5 * time.Second},
		ServerListMaxAge: Dur{1 * time.Hour},
		ServerInfoMaxAge: Dur{1 * time.Minute},
	}
	cfg, _ = def.Merge(cfg)
	return cfg
}()

func mergeVal[T comparable](p *T, v T) bool {
	var z T
	if v != z {
		*p = v
		return true
	}
	return false
}

func mergePositive[T ~int](p *T, v T) bool {
	if v > 0 {
		*p = v
		return true
	}
	return false
}

func mergeDur(p *Dur, v Dur) bool {
	if v.D > 0 {
		*p = v
		return true
	}
	return false
}

func mergeObj[T interface{ Merge(T) (T, bool) }](p *T, v T) bool {
	v, changed := (*p).Merge(v)
	if changed {
		*p = v
	}
	return changed
}

func mergeMap[T map[K]V, K comparable, V any](p *T, v T) bool {
	if v != nil {
		*p = v
		return true
	}
	return false
}

type ConnInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
}

func (c ConnInfo) Addr() string {
	h := c.Host
	if h == "" {
		h = "127.0.0.1"
	}
	return fmt.Sprintf("%s:%d", h, c.Port)
}

func (c ConnInfo) Merge(p ConnInfo) (ci ConnInfo, changed bool) {
	changed = mergeVal(&c.Host, p.Host)
	changed = mergePositive(&c.Port, p.Port) || changed
	changed = mergeVal(&c.Password, p.Password) || changed
	return c, changed
}

type Config struct {
	Netcon           ConnInfo        `json:"netcon"`
	Rcon             ConnInfo        `json:"rcon"`
	AudioDelay       Dur             `json:"audioDelay"`
	AudioLimit       Dur             `json:"audioLimit"`
	AudioLimitTTS    Dur             `json:"audioLimitTTS"`
	TextLimit        int             `json:"textLimit"`
	IncludeUsernames map[string]bool `json:"includeUsernames"`
	ExcludeUsernames map[string]bool `json:"excludeUsernames"`
	Hosts            map[string]bool `json:"hosts"`
	FirstVoice       string          `json:"firstVoice"`
	LogLevel         string          `json:"logLevel"`
	RateLimit        Dur             `json:"rateLimit"`
	ServerListMaxAge Dur             `json:"serverListMaxAge"`
	ServerInfoMaxAge Dur             `json:"serverInfoMaxAge"`

	Minimized *bool `json:"minimized"`
	Demo      *bool `json:"demo"`

	// deprecated: use netcon
	TnetPort int `json:"tnetPort"`
}

func (c Config) StartMinimized() bool {
	return c.Minimized != nil && *c.Minimized
}

func (c Config) Merge(p Config) (cfg Config, changed bool) {
	changed = mergePositive(&c.TnetPort, p.TnetPort)
	changed = mergeObj(&c.Netcon, p.Netcon) || changed
	changed = mergeObj(&c.Rcon, p.Rcon) || changed
	changed = mergeDur(&c.AudioDelay, p.AudioDelay) || changed
	changed = mergeDur(&c.AudioLimit, p.AudioLimit) || changed
	changed = mergeDur(&c.AudioLimitTTS, p.AudioLimitTTS) || changed
	changed = mergePositive(&c.TextLimit, p.TextLimit) || changed
	changed = mergeMap(&c.IncludeUsernames, p.IncludeUsernames) || changed
	changed = mergeMap(&c.ExcludeUsernames, p.ExcludeUsernames) || changed
	changed = mergeMap(&c.Hosts, p.Hosts) || changed
	changed = mergeVal(&c.LogLevel, p.LogLevel) || changed
	changed = mergeDur(&c.RateLimit, p.RateLimit) || changed
	changed = mergeDur(&c.ServerListMaxAge, p.ServerListMaxAge) || changed
	changed = mergeDur(&c.ServerInfoMaxAge, p.ServerInfoMaxAge) || changed
	changed = mergeVal(&c.Minimized, p.Minimized) || changed
	changed = mergeVal(&c.Demo, p.Demo) || changed
	return c, changed
}

func Read(fn string) (config Config, err error) {
	var cfg Config
	f, err := os.Open(fn)
	switch {
	case err == nil:
		defer f.Close()
		err = json.NewDecoder(f).Decode(&cfg)
	case os.IsNotExist(err):
		err = nil
	}
	cfg, _ = DefaultConfig.Merge(cfg)
	if err != nil {
		return cfg, fmt.Errorf("config.Read: parse: %w", err)
	}
	return cfg, nil
}

func TryRead(fn string) Config {
	cfg, _ := Read(fn)
	return cfg
}

func Write(fn string, cfg Config) error {
	s, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fn, s, 0o644)
}
