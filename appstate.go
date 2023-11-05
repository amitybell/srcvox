package main

import (
	"time"
)

var (
	DefaultAppState = AppState{
		LastUpdate: time.Now(),
		Sounds:     Sounds,
		Presence:   Presence{Error: "..."},
		Games:      GamesList,
	}
)

type AppState struct {
	LastUpdate time.Time   `json:"lastUpdate"`
	Presence   Presence    `json:"presence"`
	Sounds     []SoundInfo `json:"sounds"`
	Games      []GameInfo  `json:"games"`
	Error      AppError    `json:"error"`

	Config
}

func (s AppState) Merge(p AppState) (_ AppState, events []string) {
	if p.AudioDelay > 0 {
		s.AudioDelay = p.AudioDelay
	}
	if p.AudioLimit > 0 {
		s.AudioLimit = p.AudioLimit
	}
	if p.Sounds != nil {
		s.Sounds = p.Sounds
	}
	if p.Presence != (Presence{}) {
		events = append(events, "sv.PresenceChange")
		s.Presence = p.Presence
	}
	if p.IncludeUsernames != nil {
		s.IncludeUsernames = p.IncludeUsernames
	}
	if p.ExcludeUsernames != nil {
		s.ExcludeUsernames = p.ExcludeUsernames
	}
	if p.Error != (AppError{}) {
		s.Error = p.Error
	}
	if p.TnetPort > 0 {
		s.TnetPort = p.TnetPort
	}
	s.Config = s.Config.Merge(p.Config)
	s.LastUpdate = time.Now()
	return s, events
}
