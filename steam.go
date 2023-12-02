package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/andygrunwald/vdf"
	"github.com/fsnotify/fsnotify"
)

var (
	_ fs.FS     = (*LibFS)(nil)
	_ fs.StatFS = (*LibFS)(nil)

	libFSs = struct {
		sync.Mutex
		m map[string]*LibFS
	}{
		m: map[string]*LibFS{},
	}
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
	Dir       string
	ConfigDir string

	fs fs.FS
}

func (lib *LibFS) Open(name string) (fs.File, error) {
	return lib.fs.Open(name)
}

func (lib *LibFS) Stat(name string) (fs.FileInfo, error) {
	if sfs, ok := lib.fs.(fs.StatFS); ok {
		return sfs.Stat(name)
	}
	return os.Stat(filepath.Join(lib.Dir, filepath.FromSlash(name)))
}

func (lib *LibFS) Mtime(name string) (time.Time, error) {
	fi, err := lib.Stat(name)
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

func (lib *LibFS) watchEv(ev WatchEvent) {
	if ev.Name == lib.ConfigDir && ev.Op&(fsnotify.Create|fsnotify.Rename) != 0 {
		Watch(lib.ConfigDir)
	}
}

func (lib *LibFS) watch() {
	WatchNotify(lib.watchEv)

	if err := Watch(lib.Dir); err != nil {
		Logs.Printf("Cannot watch: `%s`: %s", lib.Dir, err)
		return
	}

	if err := Watch(lib.ConfigDir); err != nil && !errors.Is(err, fs.ErrNotExist) {
		Logs.Printf("Cannot watch: `%s`: %s", lib.ConfigDir, err)
	}
}

func NewLibFS(dir string) *LibFS {
	dir = filepath.Clean(dir)

	libFSs.Lock()
	defer libFSs.Unlock()

	if lib, ok := libFSs.m[dir]; ok {
		return lib
	}

	lib := &LibFS{
		Dir:       dir,
		ConfigDir: filepath.Join(dir, "config"),
		fs:        os.DirFS(dir),
	}
	libFSs.m[dir] = lib

	go lib.watch()

	return lib
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
	pth := "config/loginusers.vdf"
	key := "/readLoginUsers/" + lib.Dir + "/" + pth
	mtime, _ := lib.Mtime(pth)
	users, err := CacheMtime(db, mtime, key, 1, func() (accs []*SteamUser, _ error) {
		rc, err := lib.Open(pth)
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

func findSteamUser(db *DB, userID uint64) (*SteamUser, bool) {
	var users []*SteamUser
	for _, lib := range findSteamLibs() {
		a, _ := readLoginUsers(db, lib)
		users = append(users, a...)
	}

	if len(users) == 0 {
		return nil, false
	}

	var usr *SteamUser
	for _, u := range users {
		if userID != 0 && u.ID != userID {
			continue
		}

		if usr == nil || u.Ts.After(usr.Ts) {
			usr = u
		}
	}

	return usr, usr != nil
}

func openUserAvatar(db *DB, userID uint64) (fs.File, error) {
	usr, ok := findSteamUser(db, userID)
	if !ok {
		return nil, fs.ErrNotExist
	}
	return usr.Lib().Open(fmt.Sprintf("config/avatarcache/%d.png", usr.ID))
}
