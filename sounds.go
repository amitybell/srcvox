package main

import (
	"fmt"
	"github.com/amitybell/memio"
	"github.com/amitybell/piper"
	"github.com/amitybell/srcvox/files"
	"io/fs"
	"path"
)

var Sounds = func() []SoundInfo {
	fis, _ := fs.ReadDir(files.Sounds, "sounds")
	lst := make([]SoundInfo, 0, len(fis))
	for _, fi := range fis {
		fn := fi.Name()
		nm := fn[:len(fn)-len(path.Ext(fn))]
		lst = append(lst, SoundInfo{
			Name: nm,
		})
	}
	return lst
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

func SoundOrTTS(tts *piper.TTS, username, text string) (au *Audio, err error) {
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
	return au, nil
}
