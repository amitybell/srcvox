package main

type Presence struct {
	InGame      bool             `json:"inGame"`
	Error       string           `json:"error"`
	UserID      uint64           `json:"userID"`
	AvatarURL   string           `json:"avatarURL"`
	Username    string           `json:"username"`
	Clan        string           `json:"clan"`
	Name        string           `json:"name"`
	GameID      uint64           `json:"gameID"`
	GameIconURI string           `json:"gameIconURI"`
	GameHeroURI string           `json:"gameHeroURI"`
	GameDir     string           `json:"gameDir"`
	Humans      SliceSet[string] `json:"humans"`
	Bots        SliceSet[string] `json:"bots"`
}
