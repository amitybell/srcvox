package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type InGame struct {
	Error string `json:"error"`
	Count int    `json:"count"`
}

func PlayersInGame(gameID uint64) InGame {
	if Env.FakeData {
		return InGame{Count: randInt()}
	}

	// TODO: is incorrect and/or not up-to-date,
	// use https://developer.valvesoftware.com/wiki/Master_Server_Query_Protocol instead
	res, err := http.Get("https://api.steampowered.com/ISteamUserStats/GetNumberOfCurrentPlayers/v1/?appid=" + strconv.FormatUint(gameID, 10))
	if err != nil {
		return InGame{Error: err.Error()}
	}
	defer res.Body.Close()

	val := struct {
		Response struct {
			Count int `json:"player_count"`
		} `json:"response"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(&val); err != nil {
		return InGame{Error: err.Error()}
	}
	return InGame{Count: val.Response.Count}
}
