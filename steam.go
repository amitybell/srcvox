package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/amitybell/srcvox/files"

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
	ID   uint64
	Name string
	Ts   time.Time

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
	return fs.Stat(lib.fs, name)
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
	fn := filepath.Join(lib.Dir, filepath.FromSlash(pth))
	users, err := CacheMtime(db, mtime, key, 4, func() (accs []*SteamUser, _ error) {
		rc, err := lib.Open(pth)
		if err != nil {
			return nil, fmt.Errorf("readloginUsers: %s: %w", fn, err)
		}
		defer rc.Close()

		data, err := vdf.NewParser(rc).Parse()
		if err != nil {
			return nil, fmt.Errorf("readloginUsers: %s: %w", fn, err)
		}

		users, ok := data["users"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("readloginUsers: %s: %w", fn, fs.ErrNotExist)
		}

		for k, v := range users {
			id, err := strconv.ParseUint(k, 10, 64)
			if err != nil {
				Logs.Debug("cannot parse login ID",
					slog.String("fn", fn),
					slog.String("value", k),
					slog.Any("error", err))
				continue
			}
			if id == 0 {
				Logs.Debug("login ID is zero",
					slog.String("fn", fn),
					slog.String("value", k))
				continue
			}

			m, ok := v.(map[string]any)
			if !ok {
				Logs.Debug("login entry is not a map",
					slog.String("fn", fn),
					slog.Uint64("userID", id),
					slog.String("type", fmt.Sprintf("%T", v)),
					slog.String("value", fmt.Sprintf("%T", v)))
				continue
			}

			name, ok := m["PersonaName"].(string)
			if !ok {
				Logs.Debug("PersonaName is not a string",
					slog.String("fn", fn),
					slog.Uint64("userID", id),
					slog.String("type", fmt.Sprintf("%T", m["PersonaName"])),
					slog.String("value", fmt.Sprintf("%T", m["PersonaName"])))
				continue
			}

			s, ok := m["Timestamp"].(string)
			if !ok {
				Logs.Debug("Timestamp is not a string",
					slog.String("fn", fn),
					slog.Uint64("userID", id),
					slog.String("type", fmt.Sprintf("%v", m["Timestamp"])),
					slog.String("value", fmt.Sprintf("%v", m["Timestamp"])))
				continue
			}
			ts, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				Logs.Debug("cannot parse Timestamp",
					slog.String("fn", fn),
					slog.String("value", s),
					slog.Any("error", err))
				continue
			}
			if ts == 0 {
				Logs.Debug("Timestamp is zero",
					slog.String("fn", fn),
					slog.String("value", s))
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
	if err != nil && !errors.Is(err, ErrStale) {
		return users, fmt.Errorf("readloginUsers: %s: %w", fn, err)
	}
	return users, nil
}

func findSteamUser(db *DB, searchID uint64) (*SteamUser, bool) {
	var users []*SteamUser
	for _, dir := range findSteamPaths() {
		a, err := readLoginUsers(db, NewLibFS(dir))
		if err != nil {
			Logs.Debug("findSteamUser: Cannot read login", slog.Any("error", err), slog.Uint64("searchID", searchID))
		}
		users = append(users, a...)
	}

	var usr *SteamUser
	for _, u := range users {
		switch {
		case searchID != 0 && u.ID != searchID:
			Logs.Debug("findSteamUser: login skipped", slog.Uint64("searchID", searchID), slog.Uint64("loginID", u.ID))
		case usr == nil:
			usr = u
		case u.Ts.After(usr.Ts):
			usr = u
		}
	}

	if usr == nil {
		Logs.Debug("findSteamUser: No user logins found", slog.Uint64("searchID", searchID))
		return nil, false
	}
	Logs.Debug("Steam User",
		slog.Uint64("userID", usr.ID),
		slog.String("userName", usr.Name),
	)

	if Env.Demo {
		usr.Name = DemoUsername
	}

	return usr, true
}

func openUserAvatar(db *DB, userID uint64) (fs.File, error) {
	usr, ok := findSteamUser(db, userID)
	if !ok {
		return nil, fs.ErrNotExist
	}
	return usr.Lib().Open(fmt.Sprintf("config/avatarcache/%d.png", usr.ID))
}

func userAvatarURI(db *DB, userID uint64) string {
	f, err := openUserAvatar(db, userID)
	if err != nil {
		return DataURI("image/jpeg", files.DefaultAvatar)
	}
	defer f.Close()
	s, err := ReadDataURI("image/png", f)
	if err != nil {
		return DataURI("image/jpeg", files.DefaultAvatar)
	}
	return s
}

type Profile struct {
	ID        uint64 `json:"id"`
	AvatarURI string `json:"avatarURI"`
	Username  string `json:"username"`
	Clan      string `json:"clan"`
	Name      string `json:"name"`
}

func SteamProfile(db *DB, userID uint64, username string) (Profile, error) {
	if userID == 0 {
		return Profile{}, fmt.Errorf("SteamProfile: Invalid userID: %d", userID)
	}

	ttl := 2 * time.Hour
	ver := 1
	dest := fmt.Sprintf("https://steamcommunity.com/profiles/%d?xml=1", userID)
	profile, err := CacheTTL(db, ttl, dest, ver, func() (p Profile, _ error) {
		defer func() {
			p.Clan, p.Name = ClanName(p.Username)
		}()

		p = Profile{
			ID:       userID,
			Username: username,
		}

		resp, err := http.Get(dest)
		if err != nil {
			return p, err
		}
		defer resp.Body.Close()

		v := struct {
			XMLName    xml.Name `xml:"profile"`
			AvatarIcon string   `xml:"avatarIcon"`
		}{}
		if err := xml.NewDecoder(resp.Body).Decode(&v); err != nil {
			return p, err
		}

		if v.AvatarIcon != "" {
			resp, err := http.Get(v.AvatarIcon)
			if err == nil {
				defer resp.Body.Close()
				s, err := io.ReadAll(resp.Body)
				if err == nil {
					p.AvatarURI = DataURI("image/jpeg", s)
				}
			}
		}

		return p, nil
	})
	if err != nil && !errors.Is(err, ErrStale) {
		return profile, fmt.Errorf("SteamProfile: %w", err)
	}
	return profile, nil
}
