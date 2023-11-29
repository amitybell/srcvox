package main

import (
	"fmt"
	"github.com/amitybell/memio"
	"github.com/amitybell/piper"
	"github.com/amitybell/srcvox/files"
	"io/fs"
	"path"
	"sort"
)

var SoundsList, SoundsMap = func() ([]SoundInfo, map[string]SoundInfo) {
	m := map[string]SoundInfo{}

	fis, _ := fs.ReadDir(files.Sounds, "sounds")
	for _, fi := range fis {
		fn := fi.Name()
		nm := fn[:len(fn)-len(path.Ext(fn))]
		m[nm] = SoundInfo{
			Name: nm,
		}
	}

	for nm, _ := range Substites {
		m[nm] = SoundInfo{Name: nm}
	}

	l := make([]SoundInfo, 0, len(m))
	for _, si := range m {
		l = append(l, si)
	}
	sort.Slice(l, func(i, j int) bool { return l[i].Name < l[j].Name })
	return l, m
}()

type SoundInfo struct {
	Name string `json:"name"`
}

func ReadSound(name string) (*memio.File, error) {
	fn := "sounds/" + name + ".ogg"
	s, err := fs.ReadFile(files.Sounds, fn)
	if err != nil {
		return nil, fmt.Errorf("ReadSound(%s): %w", name, err)
	}
	return memio.NewFile(s), nil
}

func LoadSound(name string) (*Audio, error) {
	f, err := ReadSound(name)
	if err != nil {
		return nil, fmt.Errorf("LoadSound: %w", err)
	}
	return readAudio(name, f)
}

func SoundOrTTS(tts *piper.TTS, state AppState, username, text string) (au *Audio, err error) {
	if n := state.TextLimit; n > 0 && len(text) > n {
		text = text[:n]
	}

	_, name := ClanName(username)
	if name == "" {
		name = username
	}

	txt := Translate(name, text)
	if txt == "" {
		return nil, fmt.Errorf("SoundOrTTS(`%s`): %w", text, ErrEmptyMessage)
	}
	if au, err := LoadSound(txt); err == nil {
		return au, nil
	}
	wav, err := tts.Synthesize(txt)
	if err != nil {
		return nil, fmt.Errorf("SoundOrTTS(`%s`): Synthesize: %w", text, err)
	}
	au, err = readAudio(txt, memio.NewFile(wav))
	if err != nil {
		return nil, fmt.Errorf("SoundOrTTS(`%s`): readAudio: %w", text, err)
	}
	au.TTS = true
	return au, nil
}
