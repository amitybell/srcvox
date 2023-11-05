package main

import (
	"fmt"
	"github.com/amitybell/srcvox/files"
	"io/fs"
)

var (
	GamesList = []GameInfo{
		{
			ID:      1012110,
			Title:   "Military Conflict - Vietnam",
			DirName: "Military Conflict - Vietnam",
			IconURI: MustReadGameIconURI(1012110),
		},
	}
)

type GameInfo struct {
	ID      uint64 `json:"id"`
	Title   string `json:"title"`
	DirName string `json:"dirName"`
	IconURI string `json:"iconURI"`
}

func ReadGameIcon(id uint64) (mime string, _ []byte, _ error) {
	s, err := fs.ReadFile(files.GameIcons, fmt.Sprintf("gameicons/%d.png", id))
	if err != nil {
		return "", nil, fmt.Errorf("LoadGameIcon(%d): %w", id, err)
	}
	return "image/png", s, nil
}

func MustReadGameIconURI(id uint64) string {
	mime, s, err := ReadGameIcon(id)
	if err != nil {
		panic(err)
	}
	return DataURI(mime, s)
}
