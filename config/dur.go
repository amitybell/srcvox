package config

import (
	"encoding"
	"time"
)

var (
	_ encoding.TextMarshaler   = Dur{}
	_ encoding.TextUnmarshaler = (*Dur)(nil)
)

type Dur struct{ D time.Duration }

func (d Dur) String() string {
	return d.D.String()
}

func (d Dur) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Dur) UnmarshalText(p []byte) error {
	dur, err := time.ParseDuration(string(p))
	if err != nil {
		return err
	}
	d.D = dur
	return nil
}
