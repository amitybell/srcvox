package main

import "time"

type Presence struct {
	InGame      bool              `json:"inGame"`
	Error       string            `json:"error"`
	UserID      uint64            `json:"userID"`
	AvatarURI   string            `json:"avatarURL"`
	Username    string            `json:"username"`
	Clan        string            `json:"clan"`
	Name        string            `json:"name"`
	GameID      uint64            `json:"gameID"`
	GameIconURI string            `json:"gameIconURI"`
	GameHeroURI string            `json:"gameHeroURI"`
	GameDir     string            `json:"gameDir"`
	Humans      SliceSet[Profile] `json:"humans"`
	Bots        SliceSet[Profile] `json:"bots"`
	Server      string            `json:"server"`
	Ts          time.Time         `json:"ts"`
}
