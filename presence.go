package main

type Presence struct {
	OK       bool   `json:"ok"`
	Error    string `json:"error"`
	UserID   uint64 `json:"userID"`
	Username string `json:"username"`
	Clan     string `json:"clan"`
	Name     string `json:"name"`
	GameID   uint64 `json:"gameID"`
	IconURI  string `json:"iconURI"`
	GameDir  string `json:"gameDir"`
}
