package main

import (
	"regexp"
	"strings"
)

var (
	Translations = map[string][]string{
		"<3":  {"love"},
		"i":   {"I"},
		"im":  {"I'm"},
		"sry": {"sorry"},
		"bb":  {"bye"},
		"brb": {"back", "I'll be right back"},
		"gl":  {"good luck!"},
		"bbq": {"barbeque"},
		"hf":  {"have fun!"},
		"kfc": {"KFC", "kfc"},
		"np":  {"no problem", "no worries", "no problemo", "shit happens"},
		"ns":  {"nice shot"},
		"ko":  {"KO", "knock out"},
		"wb":  {"welcome back"},
		"icu": {"I see you!"},
		"gg":  {"good game"},
		"thx": {"thanks"},
	}

	Substites = map[string][]string{
		"ns":         {"nice", "nice shot", "noice", "goodjob"},
		"bb":         {"bye1", "bye-ni1", "bye-ni2"},
		"bye":        {"bye1", "bye-ni1", "bye-ni2"},
		"glhf":       {"good luck, have fun!"},
		"gg":         {"GG", "good game", "game", "good", "nice game", "well played"},
		"icu":        {"I see you!", "iseeyou"},
		"icq":        {"list", "look", "honey", "run2"},
		"yw":         {"you're welcome!"},
		"hacker":     {"hacker", "$name is the hacker!"},
		"hack":       {"hack", "hack the planet!"},
		"usure":      {"I'm sure"},
		"boxxy":      {"boxxy", "iamboxxy"},
		"baby":       {"corner"},
		"lol":        {"haha", "lol"},
		"ladydecade": {"ladydecade1", "ladydecade2", "ladydecade3", "ladydecade4"},
		"zombie":     {"zombie1", "zombie2"},
		"dust":       {"dust1", "dust2"},
		"drunken":    {"drunken1", "drunken2", "drunken2", "drunken4", "drunken5", "drunken6", "drunken7"},
		"run":        {"run1", "run2"},
		"shit":       {"shit1", "shit2", "shit3"},
		"THX":        {"THX"},
		"thx":        {"thanks"},
		"wololo":     {"wololo1", "wololo2"},
	}

	clanNamePat = regexp.MustCompile(`^\s*((?:\*+\s*[^*]+\*+)|(?:\[+\s*[^]]+\]+)|(?:\(+\s*[^)]+\)+))\s*(.+?)\s*$`)
)

func Translate(name, text string) string {
	if v, ok := Substites[text]; ok {
		text = randElem(v)
	} else {
		text = strings.ToLower(strings.TrimSpace(text))
		if v, ok := Substites[text]; ok {
			text = randElem(v)
		}
	}

	out := strings.Fields(text)
	for i, k := range out {
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
	if m := clanNamePat.FindStringSubmatch(username); len(m) == 3 {
		return m[1], m[2]
	}
	return "", strings.TrimSpace(username)
}
