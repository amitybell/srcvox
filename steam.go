package main

import (
	"fmt"
	"github.com/andygrunwald/vdf"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
)

type SteamUser struct {
	ID        uint64
	Name      string
	AvatarURL string
}

type LibFS struct {
	Dir string
	fs.FS
}

func findSteamPaths(rel ...string) []string {
	relFn := filepath.Join(rel...)
	paths := []string{}
	for _, dir := range steamSearchDirs {
		fn := filepath.Join(dir, relFn)
		_, err := os.Stat(fn)
		if err == nil {
			paths = append(paths, fn)
		}
	}
	return paths
}

func parseVDF(fn string) (map[string]any, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, fmt.Errorf("parseVDF: %w", err)
	}
	defer f.Close()

	p := vdf.NewParser(f)
	m, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("parseVDF: %w", err)
	}
	return m, nil
}

func findSteamLibs() []*LibFS {
	libs := []*LibFS{}

	seen := map[string]bool{}
	for _, fn := range findSteamPaths("config", "libraryfolders.vdf") {
		data, err := parseVDF(fn)
		if err != nil {
			continue
		}

		folders, ok := data["libraryfolders"].(map[string]any)
		if !ok {
			continue
		}

		for _, folder := range folders {
			lib, ok := folder.(map[string]any)
			if !ok {
				continue
			}

			dir, ok := lib["path"].(string)
			if !ok {
				continue
			}

			if seen[dir] {
				continue
			}

			seen[dir] = true
			libs = append(libs, &LibFS{Dir: dir, FS: os.DirFS(dir)})
		}
	}
	return libs
}

func readUserAccount(lib *LibFS) (*SteamUser, error) {
	rc, err := lib.Open("config/loginusers.vdf")
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := vdf.NewParser(rc).Parse()
	if err != nil {
		return nil, err
	}

	users, ok := data["users"].(map[string]any)
	if !ok {
		return nil, fs.ErrNotExist
	}

	for k, v := range users {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}

		name, ok := m["PersonaName"].(string)
		if !ok {
			continue
		}

		id, err := strconv.ParseUint(k, 10, 64)
		if err != nil || id == 0 {
			continue
		}

		usr := &SteamUser{ID: id, Name: name}
		if Env.Demo {
			usr.Name = DemoUsername
		}
		return usr, nil
	}
	return nil, fs.ErrNotExist
}

func findSteamUser() (*SteamUser, *LibFS, bool) {
	for _, lib := range findSteamLibs() {
		usr, err := readUserAccount(lib)
		if err == nil {
			return usr, lib, true
		}
	}
	return nil, nil, false
}

func openUserAvatar() (fs.File, error) {
	usr, lib, ok := findSteamUser()
	if !ok {
		return nil, fs.ErrNotExist
	}
	return lib.Open(fmt.Sprintf("config/avatarcache/%d.png", usr.ID))
}
