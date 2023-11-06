package main

type Presence struct {
	OK          bool   `json:"ok"`
	Error       string `json:"error"`
	UserID      uint64 `json:"userID"`
	Username    string `json:"username"`
	Clan        string `json:"clan"`
	Name        string `json:"name"`
	GameID      uint64 `json:"gameID"`
	GameIconURI string `json:"gameIconURI"`
	GameHeroURI string `json:"gameHeroURI"`
	GameDir     string `json:"gameDir"`
}
