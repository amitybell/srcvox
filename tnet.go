package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/amitybell/memio"
	"github.com/gopxl/beep"
	"github.com/tqwewe/go-steam/steamid"
	"github.com/ziutek/telnet"
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
	ChatPat         = regexp.MustCompile(`^(?:[*](?:DEAD|SPEC)[*])?\s*(?:\([^)]+\))?\s*(.+?)\s*:\s*(?:[#]|:\s?>|:\s?<|<\s?:|>\s?:)\s*(.+?)\s*$`)
	CvarPat         = regexp.MustCompile(`^(?:\[[^\]]+\])?"?([^"]+)"?\s*=\s*"([^"]*)"`)
	GamePathPat     = regexp.MustCompile(`(?i)^GAME\s.*"([^"]+[/\\]steamapps[/\\]common)[/\\]+([^/\\"]+)`)
	FlatpakPat      = regexp.MustCompile(`^\w+:([\\].+)`)
	ConnectPat      = regexp.MustCompile(`(?i)^\s*(.*)\s*(connected|disconnected|Not connected to server)\s*$`)
	StatusPat       = regexp.MustCompile(`^#\s+\d+(?:\s+\d+)?\s+"([^"]+)".+(STEAM_\d+:\d+:\d+|BOT)`)
	StatusServerPat = regexp.MustCompile(`^\s*Connected to (\S+:\d+)\s*$`)

	StatusTableBegin = `# userid name uniqueid connected ping loss state rate`
	StatusTableEnd   = `#end`
)

type X = []string

type Tnet struct {
	Conn *telnet.Conn
	Q    chan *Audio
	stop chan struct{}

	statusServer atomic.Pointer[string]

	app *App
}

func (tn *Tnet) Exec(cmds ...[]string) error {
	if len(cmds) == 0 {
		return nil
	}

	b := memio.NewFile(nil)
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
	cmd := b.Bytes()
	b.WriteString("\r\n")
	b.Seek(0, 0)

	if _, err := b.WriteTo(tn.Conn); err != nil {
		return fmt.Errorf("Tnet.Exec(`%s`): %w", cmd, err)
	}
	return nil
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

func (tn *Tnet) playLoop(ctx context.Context) {
	for {
		select {
		case a := <-tn.Q:
			tn.play(tn.drainQ(a))
		case <-ctx.Done():
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

func (tn *Tnet) readLineCvar(name, val string) {
	switch name {
	}
}

func (tn *Tnet) readLineGamePath(steamDir, gameNm string) {
	ts := time.Now()
	var game *GameInfo
	for _, g := range GamesList {
		if strings.EqualFold(g.DirName, gameNm) {
			game = g
			break
		}
	}
	if game == nil {
		Logs.Printf("readLineGamePath: Unsupported game: %s\n", gameNm)
		return
	}

	if m := FlatpakPat.FindStringSubmatch(steamDir); len(m) == 2 && PlatformIsLinux {
		steamDir = strings.ReplaceAll(m[1], `\`, `/`)
		if _, err := os.Stat(steamDir); err != nil {
			// TODO: replace this hack with a generic case-insensitive path resolution
			steamDir = strings.ReplaceAll(steamDir, `steam/`, `Steam/`)
		}
	}
	if _, err := os.Stat(steamDir); err != nil {
		Logs.Printf("readLineGamePath: Steam directory `%s` doesn't exist: %s\n", steamDir, err)
		return
	}

	gameDir := filepath.Join(steamDir, game.DirName)
	if _, err := os.Stat(gameDir); err != nil {
		Logs.Printf("readLineGamePath: Game directory `%s` doesn't exist: %s\n", gameDir, err)
		return
	}

	tn.app.UpdateState(func(s AppState) AppState {
		s.Presence.InGame = true
		s.Presence.GameID = game.ID
		s.Presence.GameIconURI = game.IconURI
		s.Presence.GameHeroURI = game.HeroURI
		s.Presence.GameDir = gameDir
		s.Presence.Ts = ts
		return s
	})
}

func (tn *Tnet) execStatus() error {
	return tn.Exec(X{"status"})
}

func (tn *Tnet) readStatusTable(conn *telnet.Conn) {
	ts := time.Now()
	var bots SliceSet[Profile]
	var hums SliceSet[Profile]
	for {
		ln, err := tn.Conn.ReadString('\n')
		if err != nil {
			return
		}
		ln = strings.TrimSpace(ln)
		if ln == StatusTableEnd {
			break
		}

		m := StatusPat.FindStringSubmatch(ln)
		if len(m) != 3 {
			continue
		}
		nm, id := m[1], m[2]

		if id == "BOT" {
			bots = bots.Add(Profile{Name: nm})
		} else {
			p, _ := SteamProfile(tn.app.DB, steamid.NewID(id).To64().Uint64(), nm)
			hums = hums.Add(p)
		}
	}

	tn.app.UpdateState(func(s AppState) AppState {
		s.Presence.Bots = bots
		s.Presence.Humans = hums
		s.Presence.Server = ""
		if p := tn.statusServer.Load(); p != nil {
			s.Presence.Server = *p
		}
		s.Presence.Ts = ts
		return s
	})
}

func (tn *Tnet) readLineStatusServer(addr string) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}
	addr = host + ":" + port
	tn.statusServer.Store(&addr)
}

func (tn *Tnet) readLineConnect(name, status string) {
	switch strings.ToLower(status) {
	case "connected":
		tn.execStatus()
	case "disconnected":
		tn.execStatus()
	case "not connected to server":
		tn.app.UpdateState(func(s AppState) AppState {
			s.Presence.Humans = s.Presence.Humans.Clear()
			s.Presence.Bots = s.Presence.Bots.Clear()
			return s
		})
	}
}

func (tn *Tnet) hostInGame(state AppState) (string, bool) {
	for _, p := range state.Presence.Humans.Slice() {
		if state.Hosts[p.Name] {
			return p.Name, true
		}
	}
	return "", false
}

func (tn *Tnet) ignoreChat(state AppState, name string) (reason string) {
	pr := state.Presence

	if name == pr.Username {
		return ""
	}

	if state.ExcludeUsernames[name] || state.ExcludeUsernames["*"] {
		return "excluded"
	}

	if !state.IncludeUsernames[name] && !state.IncludeUsernames["*"] {
		return "not included"
	}

	if host, ok := tn.hostInGame(state); ok {
		return "host " + host + " is in game"
	}

	if !tn.app.Limiter(name).Allow() {
		return "rate limited"
	}

	return ""
}

func (tn *Tnet) readLineChat(name, msg string) {
	state := tn.app.State()

	if r := tn.ignoreChat(state, name); r != "" {
		Logs.Printf("readLineChat: ignored: `%s: %s`: %s\n", name, msg, r)
		return
	}

	au, err := SoundOrTTS(tn.app.TTS(name), state, name, msg)
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

	if line == StatusTableBegin {
		tn.readStatusTable(tn.Conn)
		return
	}

	if ln := StatusServerPat.FindStringSubmatch(line); len(ln) == 2 {
		tn.readLineStatusServer(ln[1])
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

	if ln := ConnectPat.FindStringSubmatch(line); len(ln) == 3 {
		tn.readLineConnect(ln[1], ln[2])
		return
	}

	if ln := GamePathPat.FindStringSubmatch(line); len(ln) == 3 {
		tn.readLineGamePath(ln[1], ln[2])
		return
	}
}

func (tn *Tnet) pollHost(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			tn.execStatus()
		}
	}
}

func (tn *Tnet) commandLoop(ctx context.Context) {
	if err := tn.Exec(X{"bind", "backspace", `echo ` + StopWord}); err != nil {
		Logs.Println(err)
	}
	if err := tn.Exec(X{"path"}); err != nil {
		Logs.Println(err)
	}
	if err := tn.execStatus(); err != nil {
		Logs.Println(err)
	}

	tn.pollHost(ctx)
}

func (tn *Tnet) Loop(ctx context.Context) error {
	go tn.playLoop(ctx)
	go tn.commandLoop(ctx)

	for {
		ln, err := tn.Conn.ReadString('\n')
		if err != nil {
			return err
		}
		tn.readLine(ln)
	}
}

func retryForever[T any](ctx context.Context, maxInterval time.Duration, f func() (T, error)) (T, error) {
	interval := 100 * time.Millisecond
	for {
		r, err := f()
		if err == nil {
			return r, nil
		}
		interval = time.Duration(float64(interval) * 1.5)
		if interval > maxInterval/2 {
			interval = randRange(maxInterval/2, maxInterval)
		}
		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return r, ctx.Err()
		}
	}
}

func dialTnet(ctx context.Context, app *App) (_ *Tnet, _ context.Context, cancel func(), _ error) {
	tc, err := retryForever(ctx, 5*time.Second, func() (*telnet.Conn, error) {
		return telnet.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", app.State().TnetPort))
	})
	if err != nil {
		return nil, ctx, nil, err
	}

	ctx, cancel = context.WithCancel(ctx)
	tn := &Tnet{
		Q:    make(chan *Audio, 1<<10),
		stop: make(chan struct{}, 1),
		Conn: tc,
		app:  app,
	}
	return tn, ctx, cancel, nil
}

func tnetCleanup(app *App, retErr *error) {
	app.UpdateState(func(s AppState) AppState {
		s.Presence.Error = "disconnected"
		if retErr != nil && *retErr != nil {
			s.Presence.Error = (*retErr).Error()
		}
		s.Presence.InGame = Env.Demo
		s.Presence.Humans = s.Presence.Humans.Clear()
		s.Presence.Bots = s.Presence.Bots.Clear()
		s.Presence.Server = ""
		return s
	})
}

func startTnet(ctx context.Context, app *App) (retErr error) {
	// reset any stale data. it will be re-initialized by readLineGamePath and readLineName
	defer tnetCleanup(app, &retErr)

	tn, ctx, cancel, err := dialTnet(ctx, app)
	if err != nil {
		return err
	}
	defer cancel()

	return tn.Loop(ctx)
}

func tnet(app *App) {
	ctx := context.Background()
	for {
		startTnet(ctx, app)
		// if the server is broken (e.g. broken pipe on write), but we can connect immediately
		// we end up just burning CPU, so always wait a little before restarting
		time.Sleep(5 * time.Second)
	}
}

func recoverPanic(err *error) {
	e := recover()
	if e == nil {
		return
	}
	*err = fmt.Errorf("PANIC: %v\n%s\n", e, debug.Stack())
}

func quote(s string) string {
	for _, r := range s {
		if unicode.IsSpace(r) || !unicode.IsPrint(r) {
			return strconv.QuoteToASCII(s)
		}
	}
	return s
}
