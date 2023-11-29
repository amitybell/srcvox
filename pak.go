package main

import (
	"fmt"
	"git.lubar.me/ben/valve/vpk"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

var (
	pakFSCache = struct {
		sync.Mutex
		ents map[pakKey]pakEnt
	}{
		ents: map[pakKey]pakEnt{},
	}
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
	for _, lib := range findSteamLibs() {
		var pfs *PakFS
		if pfs, err = getPakFS(game, lib.Dir); err == nil {
			return pfs, nil
		}
	}
	return nil, fmt.Errorf("No paks found for `%s`: %v", game.Title, err)
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

func (pf *PakFile) Read(p []byte) (int, error) {
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
