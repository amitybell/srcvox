package main

type API struct {
	app *App
}

func (a *API) Log(p ...any) {
	if len(p) == 0 {
		return
	}
	// ui bindings use `any[]` instead of variadict args, so unwrap it
	if q, ok := p[0].([]any); ok && len(p) == 1 {
		p = q
	}
	Logs.Println(p...)
}

func (a *API) State() AppState {
	return a.app.State()
}

func (a *API) Sounds() []SoundInfo {
	return a.State().Sounds
}

func (a *API) InGame(gameID uint64) InGame {
	return PlayersInGame(gameID)
}

func (a *API) Games() []GameInfo {
	return a.State().Games
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

func (a *API) ServerInfos(gameID uint64) ([]ServerInfo, error) {
	return ServerInfos(a.app.DB, gameID)
}
