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
	if p.Sounds != nil {
		s.Sounds = p.Sounds
	}
	if p.Presence != (Presence{}) {
		events = append(events, "sv.PresenceChange")
		s.Presence = p.Presence
	}
	if p.Error != (AppError{}) {
		events = append(events, "sv.ErrorChange")
		s.Error = p.Error
	}
	s.Config = s.Config.Merge(p.Config)
	s.LastUpdate = time.Now()
	return s, events
}
