package main

import (
	"errors"
	"strconv"
)

var (
	ErrServerNotStarted = errors.New("Server not started")
)

type APILog struct {
	Level   string   `json:"level"`
	Message string   `json:"message"`
	Trace   []string `json:"trace"`
}

type API struct {
	app *App
}

func (a *API) Log(v APILog) {
	Logs.API(v)
}

func (a *API) State() AppState {
	return a.app.State()
}

func (a *API) Sounds() []SoundInfo {
	return SoundsList
}

func (a *API) InGame(gameID uint64) InGame {
	return PlayersInGame(gameID)
}

func (a *API) Games() []GameInfo {
	srvURL := ""
	if a.app.listener != nil {
		srvURL = "http://" + a.app.listener.Addr().String()
	}
	games := make([]GameInfo, len(GamesList))
	for i, _ := range GamesList {
		g := *GamesList[i]
		g.BgVideoURL = srvURL + "/app.bgvideo?id=" + strconv.FormatUint(g.ID, 10)
		g.MapImageURL = srvURL + "/app.mapimage?id=" + strconv.FormatUint(g.ID, 10)
		games[i] = g
	}
	return games
}

func (a *API) Error() AppError {
	return a.State().Error
}

func (a *API) Presence() Presence {
	return a.State().Presence
}

func (a *API) Env() Environment {
	return Env
}

func (a *API) Servers(gameID uint64) (map[string]Region, error) {
	return serverList(a.app.DB, gameID)
}

func (a *API) ServerInfo(region Region, addr string) (ServerInfo, error) {
	inf, _, err := serverInfo(a.app, region, addr)
	// server query might be stale
	// update it if we have more up-to-date `status` info
	p := a.app.State().Presence
	if p.Server == inf.Addr && p.Ts.After(inf.Ts) && p.Humans.Len() > 0 {
		inf.Players = p.Humans.Len()
		inf.Bots = p.Bots.Len()
	}
	return inf, err
}

func (a *API) AppAddr() (string, error) {
	lsn := a.app.listener
	if lsn == nil {
		return "", ErrServerNotStarted
	}
	return lsn.Addr().String(), nil
}

func (a *API) Profile(userID uint64, name string) (Profile, error) {
	return SteamProfile(a.app.DB, userID, name)
}
