package main

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	Translations = AltMap(map[string][]string{
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
		"np":  {"no problem", "no worries", "shit happens"},
		"ns":  {"nice shot"},
		"ko":  {"KO", "knock out"},
		"wb":  {"welcome back"},
		"icu": {"I see you!"},
		"gg":  {"good game"},
		"thx": {"thanks"},
		":)":  {""},
		":(":  {""},
		":d":  {""},
		"xd":  {""},
		"ez":  {"easy"},
		"ftw": {"for the win"},
		"jk":  {"just kidding"},
		"btw": {"by the way"},
	})

	Substites = AltMap(map[string][]string{
		"ns":              {"nice shot", "goodjob", "decent"},
		"bb":              {"bye1", "bye-ni1", "bye-ni2"},
		"bye":             {"bye1", "bye-ni1", "bye-ni2"},
		"glhf":            {"good luck, have fun!"},
		"gg":              {"GG", "good game", "game", "good", "nice game", "well played"},
		"icu":             {"I see you!", "iseeyou"},
		"icq":             {"list", "look", "honey", "run2"},
		"yw":              {"you're welcome!"},
		"hacker":          {"hacker1", "$name is the hacker!"},
		"bji, bij, bitch": {"bitch1"},
		"hack":            {"hack", "hack the planet!"},
		"usure":           {"I'm sure"},
		"boxxy":           {"isboxxy", "nottrollin", "iamboxxy"},
		"baby":            {"corner"},
		"lol":             {"haha", "lol"},
		"ladydecade":      {"ladydecade1", "ladydecade2", "ladydecade3", "ladydecade4"},
		"zombie":          {"zombie1", "zombie2"},
		"dust":            {"dust1", "dust2"},
		"drunken":         {"drunken1", "drunken2", "drunken2", "drunken4", "drunken5", "drunken6", "drunken7"},
		"run":             {"run1", "run2"},
		"shit":            {"shit1", "shit2", "shit3"},
		"THX":             {"THX"},
		"thx":             {"thanks"},
		"wololo":          {"wololo1", "wololo2"},
		"gingle":          {"jinglebell"},
		"wambulance":      {"wambulance1", "wambulance2"},
		"happynewyear":    {"newyear"},
	})

	clanNamePat = regexp.MustCompile(`^\s*((?:\*+\s*[^*]+\*+)|(?:\[+\s*[^]]+\]+)|(?:\(+\s*[^)]+\)+))\s*(.+?)\s*$`)
)

type Alt[T any] struct {
	L []T
	i int
}

func (a *Alt[T]) Next(def T) (v T) {
	if a == nil || len(a.L) == 0 {
		return def
	}
	v = a.L[a.i]
	a.i = (a.i + 1) % len(a.L)
	return v
}

func AltMap[T any](p map[string][]T) map[string]*Alt[T] {
	q := make(map[string]*Alt[T], len(p))
	for k, v := range p {
		for _, k := range strings.FieldsFunc(k, commaSpace) {
			q[k] = &Alt[T]{L: shuffle(v)}
		}
	}
	return q
}

func commaSpace(r rune) bool {
	return r == ',' || unicode.IsSpace(r)
}

func Translate(name, text string) string {
	for _, k := range []string{text, strings.ToLower(strings.TrimSpace(text))} {
		if v, ok := Substites[k]; ok {
			text = v.Next("")
			break
		}
	}

	if name == "" {
		name = "someone"
	}

	out := strings.Fields(strings.ToLower(text))
	for i, word := range out {
		switch word {
		case "$name":
			word = name
		}
		out[i] = Translations[word].Next(word)
	}

	return strings.TrimSpace(strings.Join(out, " "))
}

func ClanName(username string) (clan, name string) {
	if m := clanNamePat.FindStringSubmatch(username); len(m) == 3 {
		return m[1], m[2]
	}
	return "", strings.TrimSpace(username)
}
