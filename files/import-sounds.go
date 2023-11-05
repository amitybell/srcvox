//go:build ignore

package main

import (
	"fmt"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/wav"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func fnToName(fn string) (name, ext string) {
	name = filepath.Base(fn)
	ext = filepath.Ext(fn)
	if i := strings.IndexByte(name, '.'); i >= 0 {
		name = name[:i]
	}
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "")
	return name, ext
}

func importSound(fn string) (string, error) {
	name, _ := fnToName(fn)
	if name == "" {
		return name, fmt.Errorf("empty name for fn `%s`", fn)
	}

	inF, err := os.Open(fn)
	if err != nil {
		return name, err
	}
	defer inF.Close()

	// TODO: support other formats?
	// I've had issues with playback of other formats in the past
	// so only wav is supported for now
	inStr, inFmt, err := wav.Decode(inF)
	if err != nil {
		return name, err
	}

	outFmt := beep.Format{
		Precision:   2, // 16-bit
		NumChannels: 1, // mono
		// Valve's docs say to use 22050, but MCV appears to only support 11025
		SampleRate: 22050 / 2,
	}
	outStr := beep.Resample(15, inFmt.SampleRate, outFmt.SampleRate, inStr)

	outF, err := os.Create(filepath.Join("sounds", name+".wav"))
	if err != nil {
		return name, err
	}
	defer outF.Close()

	return name, wav.Encode(outF, outStr, outFmt)
}

func main() {
	for _, fn := range os.Args[1:] {
		status := "ok"
		name, err := importSound(fn)
		if err != nil {
			status = err.Error()
		}
		log.Println(name, status, fn)
	}
}
