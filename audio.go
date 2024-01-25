package main

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/amitybell/memio"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/vorbis"
	"github.com/gopxl/beep/wav"
)

var (
	DefaultResampleQuality = 4
)

type Audio struct {
	Name   string
	Size   int
	Dur    time.Duration
	Format beep.Format
	TTS    bool

	mu     sync.Mutex
	Stream beep.StreamSeeker
}

func (au *Audio) Encode(delay, limit time.Duration, w io.WriteSeeker, format beep.Format) (outputDuration time.Duration, err error) {
	defer recoverPanic(&err)

	au.mu.Lock()
	defer au.mu.Unlock()

	if err := au.Stream.Seek(0); err != nil {
		return 0, fmt.Errorf("Audio.Encode: audio seek: %w", err)
	}

	var stream beep.Streamer = au.Stream
	if au.Format != format {
		stream = beep.Resample(DefaultResampleQuality, au.Format.SampleRate, format.SampleRate, au.Stream)
	}

	dur := au.Dur
	if limit > 0 && dur > limit {
		dur = limit
	}

	if delay > 0 {
		stream = beep.Seq(beep.Silence(format.SampleRate.N(delay)), stream)
		dur += delay
	}

	// TODO: figure out why this break audio playback
	// explicitly limit the playback duration,
	// to avoid issues with e.g. invalid wav header data
	stream = beep.Take(format.SampleRate.N(dur), stream)

	if err := wav.Encode(w, stream, format); err != nil {
		return 0, fmt.Errorf("Audio.Encode: wav encode: %w", err)
	}

	if _, err = w.Seek(0, 0); err != nil {
		return 0, fmt.Errorf("Audio.Encode: buffer seek: %w", err)
	}
	return dur, nil
}

func (au *Audio) EncodeToFile(delay, limit time.Duration, fn string, format beep.Format) (time.Duration, error) {
	out := &memio.File{}
	dur, err := au.Encode(delay, limit, out, format)
	if err != nil {
		return 0, err
	}
	os.WriteFile(fn, out.Bytes(), 0644)
	return dur, nil
}

func decodeAudio(src *memio.File) (beep.StreamSeekCloser, beep.Format, error) {
	defer src.Seek(0, 0)

	mt := mimetype.Detect(src.Bytes())
	switch mt.String() {
	case "audio/wav":
		return wav.Decode(src)
	case "audio/ogg":
		return vorbis.Decode(src)
	case "audio/flac":
		return flac.Decode(src)
	case "audio/mpeg":
		return mp3.Decode(src)
	default:
		return nil, beep.Format{}, fmt.Errorf("Unsuppored file format: %s", mt)
	}
}

func readAudio(name string, src *memio.File) (*Audio, error) {
	stream, format, err := decodeAudio(src)
	if err != nil {
		return nil, err
	}
	sstream, ok := stream.(beep.StreamSeeker)
	if !ok {
		return nil, fmt.Errorf("Cannot get duration of `%s`: StreamSeeker implemented", name)
	}
	size := sstream.Len()
	dur := format.SampleRate.D(size)
	a := &Audio{
		Name:   name,
		Stream: sstream,
		Format: format,
		Size:   size,
		Dur:    dur,
	}
	return a, nil
}
