package main

import (
	"time"
)

const (
	SvPresenceChangeEvent = "sv.PresenceChange"
	SvErrorChangeEvent    = "sv.ErrorChange"
)

type Reducer func(p AppState) AppState

type AppState struct {
	LastUpdate time.Time `json:"lastUpdate"`
	Presence   Presence  `json:"presence"`
	Error      AppError  `json:"error"`

	Config
}

func (s AppState) Merge(p AppState) (_ AppState, events []string) {
	if p.Presence != (Presence{}) && p.Presence != s.Presence {
		events = append(events, SvPresenceChangeEvent)
		s.Presence = p.Presence
	}
	if p.Error != (AppError{}) && p.Error != s.Error {
		events = append(events, SvErrorChangeEvent)
		s.Error = p.Error
	}
	s.Config = s.Config.Merge(p.Config)
	s.LastUpdate = time.Now()
	return s, events
}
