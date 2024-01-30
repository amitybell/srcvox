package steam

type Profile struct {
	UserID    ID     `json:"userID"`
	AvatarURI string `json:"avatarURI"`
	Username  string `json:"username"`
	Clan      string `json:"clan"`
	Name      string `json:"name"`
}
