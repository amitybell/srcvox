package appstate

import (
	"time"

	"github.com/amitybell/srcvox/data"
	"github.com/amitybell/srcvox/steam"
)

type Presence struct {
	InGame      bool                         `json:"inGame"`
	Error       string                       `json:"error"`
	UserID      steam.ID                     `json:"userID"`
	AvatarURI   string                       `json:"avatarURL"`
	Username    string                       `json:"username"`
	Clan        string                       `json:"clan"`
	Name        string                       `json:"name"`
	GameID      steam.ID                     `json:"gameID"`
	GameIconURI string                       `json:"gameIconURI"`
	GameHeroURI string                       `json:"gameHeroURI"`
	GameDir     string                       `json:"gameDir"`
	Humans      data.SliceSet[steam.Profile] `json:"humans"`
	Bots        data.SliceSet[steam.Profile] `json:"bots"`
	Server      string                       `json:"server"`
	Ts          time.Time                    `json:"ts"`
}
