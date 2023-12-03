package main

import (
	"testing"
)

func TestClanName(t *testing.T) {
	cases := []struct {
		Username string
		Clan     string
		Name     string
	}{
		{"* PP * FPS!DOUG ★", "* PP *", "FPS!DOUG ★"},
		{"[PP] TEH PWNERER", "[PP]", "TEH PWNERER"},
		{"(PP) ★TAGI★", "(PP)", "★TAGI★"},
		{"KYLE", "", "KYLE"},
	}

	for _, c := range cases {
		clan, name := ClanName(c.Username)
		switch {
		case clan != c.Clan:
			t.Fatalf("Expected clan `%s`; Got `%s`", c.Clan, clan)
		case name != c.Name:
			t.Fatalf("Expected name `%s`; Got `%s`", c.Name, name)
		}
	}
}
