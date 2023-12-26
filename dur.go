package main

import (
	"encoding/json"
	"time"
)

type Dur struct{ D time.Duration }

func (d *Dur) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.D.String())
}

func (d *Dur) UnmarshalJSON(p []byte) error {
	var v string
	if err := json.Unmarshal(p, &v); err != nil {
		return err
	}

	dur, err := time.ParseDuration(v)
	if err != nil {
		return err
	}
	d.D = dur
	return nil
}
