package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gopxl/beep"
	"github.com/ziutek/telnet"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	StopWord = `#stop`
)

var (
	DefaultVoiceFormat = beep.Format{
		Precision:   2, // 16-bit
		NumChannels: 1, // mono
		// Valve docs say to use 22050, but MC:V appears to only support 11025
		SampleRate: 22050 / 2,
	}

	ChatPat     = regexp.MustCompile(`^(?:[*]DEAD[*])?\s*(?:\([^)]+\))?\s*(.+?)\s*:\s*(?:[#]|:\s?>|:\s?<|<\s?:|>\s?:)\s*(.+?)\s*$`)
	CvarPat     = regexp.MustCompile(`^"([^"]+)"\s*=\s*"([^"]*)"`)
	PathsPat    = regexp.MustCompile(`^(WORKSHOP|EXECUTABLE_PATH|PLATFORM|MOD|GAMEBIN|GAME|GAME|CONTENT|CONTENT|DEFAULT_WRITE_PATH|LOGDIR|USRLOCAL|CONFIG)\s*"([^"]*)"`)
	UserDataPat = regexp.MustCompile(`(?i)^(.+)[/\\]userdata[/\\](\d+)[/\\](\d+).+?$`)
	FlatpakPat  = regexp.MustCompile(`^\w+:([\\].+)`)
)

type X = []string

type Tnet struct {
	Conn *telnet.Conn
	Q    chan *Audio
	stop chan struct{}
	Ctx  context.Context

	app *App
}

func (t *Tnet) Exec(cmds ...[]string) error {
	if len(cmds) == 0 {
		return nil
	}

	b := bytes.NewBuffer(nil)
	for i, cmd := range cmds {
		if i > 0 {
			b.WriteString("; ")
		}
		for j, s := range cmd {
			if j > 0 {
				b.WriteString(" ")
			}
			b.WriteString(quote(s))
		}
	}
	b.WriteString("\r\n")
	_, err := t.Conn.Write(b.Bytes())
	return err
}

func (tn *Tnet) drainQ(def *Audio) *Audio {
	last := def
	for {
		select {
		case a := <-tn.Q:
			last = a
		default:
			return last
		}
	}
}

func (tn *Tnet) playLoop() {
	for {
		select {
		case a := <-tn.Q:
			a = tn.drainQ(a)
			tn.play(a)
			time.Sleep(1 * time.Second)
		case <-tn.Ctx.Done():
			return
		}
	}
}

func (tn *Tnet) play(au *Audio) (err error) {
	dir := tn.app.State().Presence.GameDir
	if dir == "" {
		return fmt.Errorf("Tnet.play: GameDir is not set")
	}
	if !filepath.IsAbs(dir) {
		return fmt.Errorf("Tnet.play: GameDir(`%s`) is not absolute", dir)
	}

	enableChat := []X{
		{`-voicerecord`},
		{`voice_scale`, `0.5`},
		{`voice_loopback`, `1`},
		{`voice_inputfromfile`, `1`},
		{`+voicerecord`},
	}
	disableChat := []X{
		{`-voicerecord`},
		{`voice_inputfromfile`, `0`},
		{`voice_loopback`, `0`},
		{`voice_scale`, `1`},
	}

	defer func() {
		e := tn.Exec(disableChat...)
		if err == nil && e != nil {
			err = e
		}
	}()

	select {
	case <-tn.stop:
		Logs.Println("voice stopped")
	default:
	}

	if err := tn.Exec(disableChat...); err != nil {
		return err
	}

	fn := filepath.Join(dir, "voice_input.wav")
	dur, err := au.EncodeToFile(tn.app.State(), fn, DefaultVoiceFormat)
	if err != nil {
		return err
	}

	if err := tn.Exec(enableChat...); err != nil {
		return err
	}

	select {
	case <-tn.stop:
	case <-time.After(dur):
	}

	return nil
}

func (tn *Tnet) readLineName(username string) {
	clan, name := ClanName(username)
	tn.app.UpdateState(func(s AppState) AppState {
		// we don't set OK, because the essectial data is set by readLineUsrLocal
		s.Presence.Username = username
		s.Presence.Clan = clan
		s.Presence.Name = name
		return s
	})
}

func (tn *Tnet) readLineCvar(name, val string) {
	switch name {
	case "name":
		tn.readLineName(val)
	}
}

func (tn *Tnet) readLineUsrLocal(val string) {
	m := UserDataPat.FindStringSubmatch(val)
	if len(m) != 4 {
		Logs.Printf("readLineUsrLocal: Failed to parse USRLOCAL path `%s`\n", val)
		return
	}
	steamDir := m[1]
	userID, err := strconv.ParseUint(m[2], 10, 64)
	if err != nil {
		Logs.Printf("readLineUsrLocal: Failed to parse USRLOCAL userID(%s): %s\n", m[2], err)
		return
	}
	gameID, err := strconv.ParseUint(m[3], 10, 64)
	if err != nil {
		Logs.Printf("readLineUsrLocal: Failed to parse USRLOCAL gameID(%s): %s\n", m[3], err)
		return
	}

	if m := FlatpakPat.FindStringSubmatch(steamDir); len(m) == 2 && runtime.GOOS == "linux" {
		steamDir = strings.ReplaceAll(m[1], `\`, `/`)
		if _, err := os.Stat(steamDir); err != nil {
			// TODO: replace this hack with a generic case-insensitive path resolution
			steamDir = strings.ReplaceAll(steamDir, `steam`, `Steam`)
		}
	}
	if _, err := os.Stat(steamDir); err != nil {
		Logs.Printf("readLineUsrLocal: Steam directory `%s` doesn't exist: %s\n", steamDir, err)
	}

	game, ok := GamesMap[gameID]
	if !ok {
		Logs.Printf("readLineUsrLocal: Unsupported gameID(%d)\n", gameID)
		return
	}

	gameDir := filepath.Join(steamDir, "steamapps", "common", game.DirName)
	if _, err := os.Stat(gameDir); err != nil {
		Logs.Printf("readLineUsrLocal: Game directory `%s` doesn't exist: %s\n", gameDir, err)
	}

	tn.app.UpdateState(func(s AppState) AppState {
		s.Presence.OK = true
		s.Presence.UserID = userID
		s.Presence.GameID = gameID
		s.Presence.GameIconURI = game.IconURI
		s.Presence.GameHeroURI = game.HeroURI
		s.Presence.GameDir = gameDir
		return s
	})
}

func (tn *Tnet) readLinePaths(name, val string) {
	switch name {
	case "USRLOCAL":
		tn.readLineUsrLocal(val)
	}
}

func (tn *Tnet) readLineChat(name, msg string) {
	state := tn.app.State()
	pr := state.Presence
	if name != pr.Username && state.ExcludeUsernames[name] {
		Logs.Println("readLineChat: excluded:", name, msg)
		return
	}
	if name != pr.Username && !state.IncludeUsernames[name] && !state.IncludeUsernames["*"] {
		Logs.Printf("readLineChat: ignored: `%s`. chatName `%s` != userName `%s`", msg, name, pr.Username)
		return
	}

	au, err := SoundOrTTS(tn.app.TTS(name), name, msg)
	if err != nil {
		Logs.Printf("Tnet.readLine: username=`%s`, message=`%s`: %s\n", name, msg, err)
		return
	}

	select {
	case tn.Q <- au:
	case <-tn.Q:
		select {
		case tn.Q <- au:
		default:
			return
		}
	default:
	}
}

func (tn *Tnet) readLine(line string) {
	line = strings.TrimSpace(line)

	if strings.ReplaceAll(line, " ", "") == StopWord {
		select {
		case tn.stop <- struct{}{}:
		default:
		}
		return
	}

	if ln := ChatPat.FindStringSubmatch(line); len(ln) == 3 {
		tn.readLineChat(ln[1], ln[2])
		return
	}

	if ln := CvarPat.FindStringSubmatch(line); len(ln) == 3 {
		tn.readLineCvar(ln[1], ln[2])
		return
	}

	if ln := PathsPat.FindStringSubmatch(line); len(ln) == 3 {
		tn.readLinePaths(ln[1], ln[2])
		return
	}
}

func (tn *Tnet) Loop() error {
	for {
		ln, err := tn.Conn.ReadString('\n')
		if err != nil {
			return err
		}
		tn.readLine(ln)
	}
}

func dialTnet(ctx context.Context, app *App) (_ *Tnet, cancel func(), _ error) {
	tc, err := telnet.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", app.State().TnetPort))
	if err != nil {
		return nil, nil, err
	}
	ctx, cancel = context.WithCancel(ctx)
	tn := &Tnet{
		Q:    make(chan *Audio, 1<<10),
		stop: make(chan struct{}, 1),
		Conn: tc,
		Ctx:  ctx,
		app:  app,
	}
	return tn, cancel, nil
}

func startTnet(ctx context.Context, app *App) error {
	tn, cancel, err := dialTnet(ctx, app)
	if err != nil {
		app.Update(AppState{Presence: Presence{Error: err.Error()}})
		return err
	}
	defer cancel()

	// reset any stale data. it will be re-initialized by readLineUsrLocal and readLineName
	app.Update(AppState{Presence: Presence{}})
	defer func() { app.Update(AppState{Presence: Presence{Error: "disconnected"}}) }()

	tn.Exec(X{"bind", "backspace", `echo ` + StopWord})
	tn.Exec(X{"name"})
	tn.Exec(X{"path"})
	go tn.playLoop()
	return tn.Loop()
}

func tnet(app *App) {
	ctx := context.Background()
	for {
		startTnet(ctx, app)
		time.Sleep(5 * time.Second)
	}
}

func recoverPanic(err *error) {
	e := recover()
	if e == nil {
		return
	}
	*err = fmt.Errorf("PANIC: %v\n%s\n", e, debug.Stack())
	Logs.Println(*err)
}

func quote(s string) string {
	for _, r := range s {
		if unicode.IsSpace(r) || !unicode.IsPrint(r) {
			return strconv.QuoteToASCII(s)
		}
	}
	return s
}
