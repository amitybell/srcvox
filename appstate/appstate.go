package appstate

import (
	"time"

	"github.com/amitybell/srcvox/config"
)

const (
	SvPresenceChangeEvent = "sv.PresenceChange"
	SvConfigChangeEvent   = "sv.ConfigChange"
	SvErrorChangeEvent    = "sv.ErrorChange"
	SvServerInfoChange    = "sv.ServerInfoChange"
)

type Reducer func(p AppState) AppState

type AppState struct {
	LastUpdate time.Time `json:"lastUpdate"`
	Presence   Presence  `json:"presence"`
	Error      AppError  `json:"error"`

	config.Config
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
	cfgChanged := false
	s.Config, cfgChanged = s.Config.Merge(p.Config)
	if cfgChanged {
		events = append(events, SvConfigChangeEvent)
	}
	if cfgChanged && len(events) != 0 {
		s.LastUpdate = time.Now()
	}
	return s, events
}
