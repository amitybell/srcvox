package steam

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

	"github.com/amitybell/srcvox/data"
	"github.com/amitybell/srcvox/demo"
	"github.com/amitybell/srcvox/files"
	"github.com/amitybell/srcvox/logs"
	"github.com/amitybell/srcvox/store"
	"github.com/amitybell/srcvox/translate"
	"github.com/amitybell/srcvox/watch"

	"github.com/andygrunwald/vdf"
	"github.com/fsnotify/fsnotify"
)

var (
	Logs = logs.AppLogger()

	_ fs.FS     = (*LibFS)(nil)
	_ fs.StatFS = (*LibFS)(nil)

	libFSs = struct {
		sync.Mutex
		m map[string]*LibFS
	}{
		m: map[string]*LibFS{},
	}
)

type User struct {
	ID   ID
	Name string
	Ts   time.Time

	lib *LibFS `msgpack:"-"`
}

func (su *User) String() string {
	return fmt.Sprintf("%s#%s", su.Name, su.ID)
}

func (su *User) Lib() *LibFS {
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

func (lib *LibFS) watchEv(ev watch.Event) {
	if ev.Name == lib.ConfigDir && ev.Op&(fsnotify.Create|fsnotify.Rename) != 0 {
		watch.Path(lib.ConfigDir)
	}
}

func (lib *LibFS) watch() {
	watch.Notify(lib.watchEv)

	if err := watch.Path(lib.Dir); err != nil {
		Logs.Printf("Cannot watch: `%s`: %s", lib.Dir, err)
		return
	}

	if err := watch.Path(lib.ConfigDir); err != nil && !errors.Is(err, fs.ErrNotExist) {
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

func FindSteamPaths(rel ...string) []string {
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

func FindSteamLibs() []*LibFS {
	libs := []*LibFS{}

	seen := map[string]bool{}
	for _, fn := range FindSteamPaths("config", "libraryfolders.vdf") {
		type Data struct {
			LibraryFolders map[string]struct {
				Path string
			}
		}

		data, err := ReadVDF[Data](fn)
		if err != nil {
			Logs.Debug("Cannot read library folders", slog.String("fn", fn), slog.String("err", err.Error()))
			continue
		}
		for _, lib := range data.LibraryFolders {
			if seen[lib.Path] {
				continue
			}
			seen[lib.Path] = true
			libs = append(libs, NewLibFS(lib.Path))
		}
	}
	return libs
}

func ReadLoginUsers(db *store.DB, lib *LibFS) ([]*User, error) {
	pth := "config/loginusers.vdf"
	key := "/ReadLoginUsers/" + lib.Dir + "/" + pth
	mtime, _ := lib.Mtime(pth)
	fn := filepath.Join(lib.Dir, filepath.FromSlash(pth))
	ver := 8
	users, err := store.CacheMtime(db, mtime, key, ver, func() (accs []*User, _ error) {
		rc, err := lib.Open(pth)
		if err != nil {
			return nil, fmt.Errorf("ReadloginUsers: %s: %w", fn, err)
		}
		defer rc.Close()

		data, err := vdf.NewParser(rc).Parse()
		if err != nil {
			return nil, fmt.Errorf("ReadloginUsers: %s: %w", fn, err)
		}

		users, ok := data["users"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("ReadloginUsers: %s: %w", fn, fs.ErrNotExist)
		}

		for k, v := range users {
			id, err := ParseID(k)
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
					slog.String("userID", id.String()),
					slog.String("type", fmt.Sprintf("%T", v)),
					slog.String("value", fmt.Sprintf("%T", v)))
				continue
			}

			name, ok := m["PersonaName"].(string)
			if !ok {
				Logs.Debug("PersonaName is not a string",
					slog.String("fn", fn),
					slog.String("userID", id.String()),
					slog.String("type", fmt.Sprintf("%T", m["PersonaName"])),
					slog.String("value", fmt.Sprintf("%T", m["PersonaName"])))
				continue
			}

			s, ok := m["Timestamp"].(string)
			if !ok {
				Logs.Debug("Timestamp is not a string",
					slog.String("fn", fn),
					slog.String("userID", id.String()),
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

			accs = append(accs, &User{
				ID:   id,
				Name: name,
				Ts:   time.Unix(ts, 0),
				lib:  lib,
			})
		}
		return accs, nil
	})
	for i := range users {
		users[i].lib = lib
	}
	if err != nil && !errors.Is(err, store.ErrStale) {
		return users, fmt.Errorf("ReadloginUsers: %s: %w", fn, err)
	}
	return users, nil
}

func FindUser(db *store.DB, searchID ID) (*User, bool) {
	var users []*User
	for _, dir := range FindSteamPaths() {
		a, err := ReadLoginUsers(db, NewLibFS(dir))
		if err != nil {
			Logs.Debug("steam.FindUser: Cannot read login", slog.Any("error", err), slog.String("searchID", searchID.String()))
		}
		users = append(users, a...)
	}

	var usr *User
	for _, u := range users {
		switch {
		case searchID != 0 && u.ID != searchID:
			Logs.Debug("steam.FindUser: login skipped", slog.String("searchID", searchID.String()), slog.String("loginID", u.ID.String()), slog.String("name", u.Name))
		case usr == nil:
			usr = u
		case u.Ts.After(usr.Ts):
			usr = u
		}
	}

	if usr == nil {
		Logs.Debug("steam.FindUser: No user logins found", slog.String("searchID", searchID.String()))
		return nil, false
	}
	Logs.Debug("Steam User",
		slog.String("userID", usr.ID.String()),
		slog.String("name", usr.Name),
		slog.Uint64("id64", usr.ID.To64()),
		slog.Uint64("id32", uint64(usr.ID.To32())),
	)

	if demo.Enabled {
		usr.Name = demo.Username
	}

	return usr, true
}

func OpenAvatar(db *store.DB, userID ID) (fs.File, error) {
	usr, ok := FindUser(db, userID)
	if !ok {
		return nil, fs.ErrNotExist
	}
	return usr.Lib().Open(fmt.Sprintf("config/avatarcache/%d.png", usr.ID.To64()))
}

func AvatarURI(db *store.DB, userID ID) string {
	f, err := OpenAvatar(db, userID)
	if err != nil {
		return data.URI("image/jpeg", files.DefaultAvatar)
	}
	defer f.Close()
	s, err := data.ReadURI("image/png", f)
	if err != nil {
		return data.URI("image/jpeg", files.DefaultAvatar)
	}
	return s
}

func FetchProfile(db *store.DB, userID ID, username string) (Profile, error) {
	if userID == 0 {
		return Profile{}, fmt.Errorf("SteamProfile: Invalid userID: %s", userID)
	}

	ttl := 2 * time.Hour
	ver := 5
	dest := fmt.Sprintf("https://steamcommunity.com/profiles/%d?xml=1", userID.To64())
	profile, err := store.CacheTTL(db, ttl, dest, ver, func() (p Profile, _ error) {
		defer func() {
			p.Clan, p.Name = translate.ClanName(p.Username)
		}()

		p = Profile{
			UserID:   userID,
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
					p.AvatarURI = data.URI("image/jpeg", s)
				}
			}
		}

		return p, nil
	})
	if err != nil && !errors.Is(err, store.ErrStale) {
		return profile, fmt.Errorf("SteamProfile: %w", err)
	}
	return profile, nil
}
