package main

import (
	"fmt"
	"io/fs"
	"path"
	"strconv"

	"github.com/amitybell/srcvox/files"
)

var (
	GamesList = []*GameInfo{
		MustInitGameInfo(GameInfo{
			ID:          1012110,
			Title:       "Military Conflict - Vietnam",
			MapImageDir: "materials/panorama/images/map_icons/playmenu",
			BgVideoFn:   "panorama/videos/background.webm",
			PakDir:      "vietnam",
			PakPfx:      "vietnam/pak01",
		}),
	}

	GamesMap = func() map[uint64]*GameInfo {
		m := map[uint64]*GameInfo{}
		for _, g := range GamesList {
			m[g.ID] = g
		}
		return m
	}()

	GamesMapString = func() map[string]*GameInfo {
		m := map[string]*GameInfo{}
		for _, g := range GamesList {
			m[strconv.FormatUint(g.ID, 10)] = g
		}
		return m
	}()
)

func (g *GameInfo) OpenFile(path string) (fs.File, error) {
	pfs, err := GetPakFS(g)
	if err != nil {
		return nil, err
	}
	return pfs.Open(path)
}

func (g *GameInfo) OpenBgVideo() (fs.File, error) {
	return g.OpenFile(g.BgVideoFn)
}

func (g *GameInfo) OpenMapImage(name string) (fs.File, error) {
	return g.OpenFile(path.Join(g.MapImageDir, name+".png"))
}

func (g *GameInfo) ReadMapNames() ([]string, error) {
	pfs, err := GetPakFS(g)
	if err != nil {
		return nil, err
	}

	names := []string{}
	ents, err := fs.ReadDir(pfs, g.MapImageDir)
	if err != nil && len(ents) == 0 {
		return nil, err
	}

	for _, de := range ents {
		nm := path.Base(de.Name())
		ext := path.Ext(nm)
		if ext == ".png" {
			names = append(names, nm[:len(nm)-len(ext)])
		}
	}
	return names, nil
}

type GameImageKind string

const (
	IconImage GameImageKind = "icon"
	HeroImage GameImageKind = "library_hero"
)

type GameInfo struct {
	ID           uint64 `json:"id"`
	Title        string `json:"title"`
	DirName      string `json:"dirName"`
	IconURI      string `json:"iconURI"`
	HeroURI      string `json:"heroURI"`
	MapImageDir  string
	MapImageURL  string   `json:"mapImageURL"`
	BgVideoURL   string   `json:"bgVideoURL"`
	MapNames     []string `json:"mapNames"`
	MapImageURLs []string `json:"mapImageURLs"`
	BgVideoFn    string
	PakDir       string
	PakPfx       string
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

func MustInitGameInfo(g GameInfo) *GameInfo {
	if g.ID == 0 {
		panic("ID is not set")
	}
	if g.Title == "" {
		panic("Title is not set")
	}
	if g.DirName == "" {
		g.DirName = g.Title
	}
	if g.IconURI == "" {
		g.IconURI = MustReadGameImageURI(g.ID, IconImage)
	}
	if g.HeroURI == "" {
		g.HeroURI = MustReadGameImageURI(g.ID, HeroImage)
	}
	return &g
}
