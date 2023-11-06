package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	Translations = map[string][]string{
		"<3":   {"love"},
		"i":    {"I"},
		"sry":  {"sorry"},
		"bb":   {"bye", "bye bye", "see ya"},
		"gl":   {"good luck!"},
		"bbq":  {"barbeque"},
		"hf":   {"have fun!"},
		"kfc":  {"KFC"},
		"gg":   {"GG", "good game"},
		"np":   {"no problem", "no worries", "no problemo", "shit happens"},
		"glhf": {"good luck, have fun!"},
		"ns":   {"nice", "nice shot", "noice"},
		"ko":   {"KO", "knock out"},
		"wb":   {"welcome back"},
	}

	Substites = map[string][]string{
		"icu":    {"I see you!"},
		"icq":    {"I seek you!", "list", "look", "you just made the list"},
		"yw":     {"you're welcome!"},
		"zombie": {"zombie", "drunken master"},
		"hacker": {"hacker", "$name is the hacker!"},
		"hack":   {"hack", "hack the planet!"},
		"usure":  {"I'm sure"},
	}
)

func Translate(name, text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	if v, ok := Substites[text]; ok {
		text = randElem(v)
	}

	out := strings.Fields(text)
	for i, k := range out {
		k := strings.ToLower(k)
		switch k {
		case "$name":
			if name != "" {
				k = name
			} else {
				k = "someone"
			}
		}

		v, _ := Translations[k]
		switch len(v) {
		case 0:
			out[i] = k
		case 1:
			out[i] = v[0]
		default:
			out[i] = randElem(v)
		}
	}

	return strings.Join(out, " ")
}

func ClanName(username string) (clan, name string) {
	i := strings.LastIndexFunc(username, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsSpace(r) && r != '-' && r != '_'
	})
	if i < 0 {
		return "", strings.TrimSpace(username)
	}
	_, n := utf8.DecodeRuneInString(username[i:])
	i += n
	return strings.TrimSpace(username[:i]), strings.TrimSpace(username[i:])
}
