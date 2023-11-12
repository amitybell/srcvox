package main

import (
	"fmt"
	"github.com/amitybell/srcvox/files"
	"io/fs"
)

var (
	GamesList = []GameInfo{
		MustInitGameInfo(1012110, "Military Conflict - Vietnam"),
	}

	GamesMap = func() map[uint64]GameInfo {
		m := map[uint64]GameInfo{}
		for _, g := range GamesList {
			m[g.ID] = g
		}
		return m
	}()
)

type GameImageKind string

const (
	IconImage GameImageKind = "icon"
	HeroImage GameImageKind = "library_hero"
)

type GameInfo struct {
	ID      uint64 `json:"id"`
	Title   string `json:"title"`
	DirName string `json:"dirName"`
	IconURI string `json:"iconURI"`
	HeroURI string `json:"heroURI"`
}

func ReadGameImage(id uint64, kind GameImageKind) (mime string, _ []byte, _ error) {
	s, err := fs.ReadFile(files.Games, fmt.Sprintf("games/%d_%s.jpg", id, kind))
	if err != nil {
		return "", nil, fmt.Errorf("LoadGameIcon(%d): %w", id, err)
	}
	return "image/jpeg", s, nil
}

func MustReadGameImageURI(id uint64, kind GameImageKind) string {
	mime, s, err := ReadGameImage(id, kind)
	if err != nil {
		panic(err)
	}
	return DataURI(mime, s)
}

func MustInitGameInfo(id uint64, name string) GameInfo {
	return GameInfo{
		ID:      id,
		Title:   name,
		DirName: name,
		IconURI: MustReadGameImageURI(id, IconImage),
		HeroURI: MustReadGameImageURI(id, HeroImage),
	}
}
