package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/andygrunwald/vdf"
)

type SteamUser struct {
	ID        uint64
	Name      string
	AvatarURL string
	Ts        time.Time

	lib *LibFS `msgpack:"-"`
}

func (su *SteamUser) String() string {
	return fmt.Sprintf("%s#%d", su.Name, su.ID)
}

func (su *SteamUser) Lib() *LibFS {
	return su.lib
}

type LibFS struct {
	Dir string

	fs fs.FS
}

func (lfs *LibFS) Open(name string) (fs.File, error) {
	return lfs.fs.Open(name)
}

func NewLibFS(dir string) *LibFS {
	return &LibFS{Dir: dir, fs: os.DirFS(dir)}
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
			libs = append(libs, NewLibFS(dir))
		}
	}
	return libs
}

func readLoginUsers(db *DB, lib *LibFS) ([]*SteamUser, error) {
	rc, err := lib.Open("config/loginusers.vdf")
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	users, err := CacheStat(db, rc, "/readLoginUsers/"+lib.Dir, 1, func() (accs []*SteamUser, _ error) {
		data, err := vdf.NewParser(rc).Parse()
		if err != nil {
			return nil, err
		}

		users, ok := data["users"].(map[string]any)
		if !ok {
			return nil, fs.ErrNotExist
		}

		for k, v := range users {
			id, err := strconv.ParseUint(k, 10, 64)
			if err != nil || id == 0 {
				continue
			}

			m, ok := v.(map[string]any)
			if !ok {
				continue
			}

			name, ok := m["PersonaName"].(string)
			if !ok {
				continue
			}

			s, ok := m["Timestamp"].(string)
			if !ok {
				continue
			}
			ts, err := strconv.ParseInt(s, 10, 64)
			if err != nil || id == 0 {
				continue
			}

			accs = append(accs, &SteamUser{
				ID:   id,
				Name: name,
				Ts:   time.Unix(ts, 0),
				lib:  lib,
			})
		}
		return accs, nil
	})
	for i, _ := range users {
		users[i].lib = lib
	}
	return users, err
}

func findSteamUser(db *DB) (*SteamUser, bool) {
	var users []*SteamUser
	for _, lib := range findSteamLibs() {
		accs, _ := readLoginUsers(db, lib)
		users = append(users, accs...)
	}

	if len(users) == 0 {
		return nil, false
	}

	sort.Slice(users, func(i, j int) bool {
		return users[j].Ts.Before(users[i].Ts)
	})

	return users[0], true
}

func openUserAvatar(db *DB) (fs.File, error) {
	usr, ok := findSteamUser(db)
	if !ok {
		return nil, fs.ErrNotExist
	}
	return usr.Lib().Open(fmt.Sprintf("config/avatarcache/%d.png", usr.ID))
}
