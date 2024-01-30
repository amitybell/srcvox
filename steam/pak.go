package steam

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"git.lubar.me/ben/valve/vpk"
	"github.com/amitybell/srcvox/demo"
	"github.com/amitybell/srcvox/errs"
)

var (
	pakFSCache = struct {
		sync.Mutex
		ents map[pakKey]pakEnt
	}{
		ents: map[pakKey]pakEnt{},
	}

	_ fs.ReadDirFS = (*PakFS)(nil)
)

func getPakFS(game *GameInfo, libDir string) (*PakFS, error) {
	gameDir := filepath.Join(libDir, "steamapps", "common", game.DirName)
	k := pakKey{
		Dir: filepath.Join(gameDir, filepath.FromSlash(game.PakDir)),
		Pfx: filepath.Join(gameDir, filepath.FromSlash(game.PakPfx)),
	}

	pakFSCache.Lock()
	defer pakFSCache.Unlock()

	if e, ok := pakFSCache.ents[k]; ok {
		return e.FS, e.Err
	}

	pfs, err := NewPakFS(k.Dir, k.Pfx)
	if err != nil {
		return nil, err
	}

	pakFSCache.ents[k] = pakEnt{FS: pfs}
	return pfs, nil
}

func GetPakFS(game *GameInfo) (*PakFS, error) {
	var err error
	for _, lib := range FindSteamLibs() {
		var pfs *PakFS
		if pfs, err = getPakFS(game, lib.Dir); err == nil {
			return pfs, nil
		}
	}
	return nil, fmt.Errorf("No paks found for `%s`: %v", game.Title, err)
}

type PakDirEntry struct {
	Fn   string
	Mode fs.FileMode
	Inf  fs.FileInfo
}

func (p *PakDirEntry) Name() string { return p.Fn }

func (p *PakDirEntry) IsDir() bool { return p.Type().IsDir() }

func (p *PakDirEntry) Type() fs.FileMode {
	if p.Mode == 0 {
		return fs.ModeIrregular
	}
	return p.Mode
}

func (p *PakDirEntry) Info() (fs.FileInfo, error) {
	if p.Inf == nil {
		return nil, fs.ErrNotExist
	}
	return p.Inf, nil
}

type pakKey struct {
	Dir string
	Pfx string
}

type pakEnt struct {
	FS  *PakFS
	Err error
}

type PakFile struct {
	nm string
	rc io.ReadCloser
}

func (pf *PakFile) Name() string {
	return pf.nm
}

func (pf *PakFile) String() string {
	return pf.Name()
}

func (pf *PakFile) Read(p []byte) (n int, err error) {
	defer errs.Recover(&err)
	return pf.rc.Read(p)
}

func (pf *PakFile) Close() error {
	return pf.rc.Close()
}

func (pf *PakFile) Stat() (fs.FileInfo, error) {
	if st, ok := pf.rc.(interface{ Stat() (fs.FileInfo, error) }); ok {
		return st.Stat()
	}
	return nil, fs.ErrInvalid
}

type PakFS struct {
	Dir string
	Pfx string

	opn  *vpk.Opener
	arc  *vpk.Archive
	ents map[string]vpk.File
}

func (p *PakFS) openOS(name string) (*PakFile, error) {
	if p.Dir == "" {
		return nil, fs.ErrNotExist
	}
	fn := filepath.Join(p.Dir, filepath.FromSlash(name))
	f, err := os.Open(fn)
	if err != nil {
		return nil, fmt.Errorf("PakFS.Open: %s: %w", name, err)
	}
	return &PakFile{nm: name, rc: f}, nil
}

func (p *PakFS) openPak(name string) (*PakFile, error) {
	f, ok := p.ents[name]
	if !ok {
		return nil, fmt.Errorf("PakFS.Open: lookup: %s: %w", name, fs.ErrNotExist)
	}

	rc, err := f.Open(p.opn)
	if err != nil {
		return nil, fmt.Errorf("PakFS.Open: open file: %s: %w", name, err)
	}

	return &PakFile{nm: name, rc: rc}, nil
}

func (p *PakFS) Open(name string) (fs.File, error) {
	name = filepath.ToSlash(name)

	if f, err := p.openOS(name); err == nil {
		return f, nil
	}
	return p.openPak(name)
}

func (p *PakFS) readDirOS(name string) ([]fs.DirEntry, error) {
	if demo.Enabled {
		return nil, nil
	}
	if p.Dir == "" {
		return nil, fs.ErrNotExist
	}
	fn := filepath.Join(p.Dir, filepath.FromSlash(name))
	return os.ReadDir(fn)
}

func (p *PakFS) readDirPak(name string) ([]fs.DirEntry, error) {
	pfx := filepath.ToSlash(name) + "/"
	ents := make([]fs.DirEntry, 0, len(p.ents))
	for _, f := range p.ents {
		if !strings.HasPrefix(f.Name(), pfx) {
			continue
		}
		ents = append(ents, &PakDirEntry{
			Fn: f.Name(),
		})
	}
	return ents, nil
}

func (p *PakFS) ReadDir(name string) ([]fs.DirEntry, error) {
	name = filepath.ToSlash(name)
	seen := map[string]bool{}
	var ents []fs.DirEntry

	osEnts, _ := p.readDirOS(name)
	pakEnts, _ := p.readDirPak(name)
	for _, dl := range [][]fs.DirEntry{osEnts, pakEnts} {
		for _, de := range dl {
			if seen[de.Name()] {
				continue
			}
			seen[de.Name()] = true
			ents = append(ents, de)
		}
	}
	return ents, nil
}

func NewPakFS(dir, pfx string) (*PakFS, error) {
	opn := vpk.Dir(pfx)
	arc, err := opn.ReadArchive()
	if err != nil {
		return nil, fmt.Errorf("NewPakFS: read archive: %w", err)
	}
	pfs := &PakFS{
		Dir:  dir,
		Pfx:  pfx,
		opn:  opn,
		arc:  arc,
		ents: make(map[string]vpk.File, len(arc.Files)),
	}
	for _, f := range arc.Files {
		pfs.ents[f.Name()] = f
	}
	return pfs, nil
}
