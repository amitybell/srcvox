package main

import (
	"errors"
	"fmt"

	"github.com/amitybell/srcvox/appstate"
	"github.com/amitybell/srcvox/config"
	"github.com/amitybell/srcvox/logs"
	"github.com/amitybell/srcvox/sound"
	"github.com/amitybell/srcvox/steam"
)

var (
	ErrServerNotStarted = errors.New("Server not started")
)

type API struct {
	app *App
}

func (a *API) Log(v logs.APILog) {
	Logs.API(v)
}

func (a *API) State() appstate.AppState {
	return a.app.State()
}

func (a *API) Config() config.Config {
	return a.app.State().Config
}

func (a *API) UpdateConfig(c config.Config) error {
	return a.app.UpdateConfig(c)
}

func (a *API) Sounds() []sound.SoundInfo {
	return sound.SoundsList
}

func (a *API) Games() []steam.GameInfo {
	srvURL := ""
	if a.app.listener != nil {
		srvURL = "http://" + a.app.listener.Addr().String()
	}
	games := make([]steam.GameInfo, len(steam.GamesList))
	for i := range steam.GamesList {
		g := *steam.GamesList[i]
		g.BgVideoURL = fmt.Sprintf("%s/app.bgvideo?id=%s", srvURL, g.ID)
		g.MapImageURL = fmt.Sprintf("%s/app.mapimage?id=%s", srvURL, g.ID)
		g.MapNames, _ = g.ReadMapNames()
		g.MapImageURLs = make([]string, len(g.MapNames))
		for i, nm := range g.MapNames {
			g.MapImageURLs[i] = fmt.Sprintf("%s/app.mapimage?id=%s&map=%s", srvURL, g.ID, nm)
		}
		games[i] = g
	}
	return games
}

func (a *API) LaunchOptions(userID, gameID steam.ID) string {
	cfg, _ := steam.ReadLocalConfig(userID)
	return cfg.Apps[gameID].LaunchOptions
}

func (a *API) Error() appstate.AppError {
	return a.State().Error
}

func (a *API) Presence() appstate.Presence {
	return a.State().Presence
}

func (a *API) Servers(gameID steam.ID) (map[string]steam.Region, error) {
	state := a.app.State()
	return steam.QueryServerList(a.app.DB, state.ServerListMaxAge.D, gameID)
}

func (a *API) ServerInfo(region steam.Region, addr string) (steam.ServerInfo, error) {
	state := a.app.State()
	inf, _, err := steam.QueryServerInfo(a.app.DB, state.ServerInfoMaxAge.D, region, addr)
	// server query might be stale
	// update it if we have more up-to-date `status` info
	p := state.Presence
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

func (a *API) Profile(userID steam.ID, name string) (steam.Profile, error) {
	return steam.FetchProfile(a.app.DB, userID, name)
}
