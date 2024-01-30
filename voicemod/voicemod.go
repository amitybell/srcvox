package voicemod

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/amitybell/memio"
	"github.com/amitybell/piper"
	"github.com/amitybell/srcvox/appstate"
	"github.com/amitybell/srcvox/audio"
	"github.com/amitybell/srcvox/data"
	"github.com/amitybell/srcvox/logs"
	"github.com/amitybell/srcvox/platform"
	"github.com/amitybell/srcvox/rng"
	"github.com/amitybell/srcvox/sound"
	"github.com/amitybell/srcvox/steam"
	"github.com/amitybell/srcvox/store"
	"github.com/gopxl/beep"
	"github.com/ziutek/telnet"
	"golang.org/x/time/rate"
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

	passwordRequiredMsg  = "must send pass command"
	passwordIncorrectMsg = "bad password attempt"

	ErrPassword = errors.New("Password is incorrect and/or not set")
)

type App interface {
	Store() *store.DB
	State() appstate.AppState
	Logs() *logs.Logger
	Limiter(name string) *rate.Limiter
	TTS(key string) *piper.TTS
	VoiceModStopped(err error)
	VoiceModGame(ts time.Time, game *steam.GameInfo, gameDir string)
	VoiceModPresence(ts time.Time, server string, hums, bots data.SliceSet[steam.Profile])
	VoiceModServerDisconnected()
}

type X = []string

type voiceMod struct {
	Logs *logs.Logger
	Conn *telnet.Conn
	Q    chan *audio.Audio
	stop chan struct{}

	statusServer atomic.Pointer[string]

	app App
}

func (vm *voiceMod) Exec(cmds ...[]string) error {
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

	if _, err := b.WriteTo(vm.Conn); err != nil {
		return fmt.Errorf("voiceMod.Exec(`%s`): %w", cmd, err)
	}
	return nil
}

func (vm *voiceMod) drainQ(def *audio.Audio) *audio.Audio {
	last := def
	for {
		select {
		case a := <-vm.Q:
			last = a
		default:
			return last
		}
	}
}

func (vm *voiceMod) playLoop(ctx context.Context) {
	for {
		select {
		case a := <-vm.Q:
			vm.play(vm.drainQ(a))
		case <-ctx.Done():
			return
		}
	}
}

func (vm *voiceMod) play(au *audio.Audio) (err error) {
	state := vm.app.State()
	dir := state.Presence.GameDir
	if dir == "" {
		return fmt.Errorf("voiceMod.play: GameDir is not set")
	}
	if !filepath.IsAbs(dir) {
		return fmt.Errorf("voiceMod.play: GameDir(`%s`) is not absolute", dir)
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
		e := vm.Exec(disableChat...)
		if err == nil && e != nil {
			err = e
		}
	}()

	select {
	case <-vm.stop:
		vm.Logs.Println("voice stopped")
	default:
	}

	if err := vm.Exec(disableChat...); err != nil {
		return err
	}

	delay := state.AudioDelay.D
	limit := state.AudioLimit.D
	if au.TTS {
		limit = state.AudioLimitTTS.D
	}

	fn := filepath.Join(dir, "voice_input.wav")
	dur, err := au.EncodeToFile(delay, limit, fn, DefaultVoiceFormat)
	if err != nil {
		return err
	}

	if err := vm.Exec(enableChat...); err != nil {
		return err
	}

	select {
	case <-vm.stop:
	case <-time.After(dur):
	}

	return nil
}

func (vm *voiceMod) readLineCvar(name, val string) {
	switch name {
	}
}

func (vm *voiceMod) readLineGamePath(steamDir, gameNm string) {
	ts := time.Now()
	var game *steam.GameInfo
	for _, g := range steam.GamesList {
		if strings.EqualFold(g.DirName, gameNm) {
			game = g
			break
		}
	}
	if game == nil {
		vm.Logs.Printf("readLineGamePath: Unsupported game: %s\n", gameNm)
		return
	}

	if m := FlatpakPat.FindStringSubmatch(steamDir); len(m) == 2 && platform.IsLinux {
		steamDir = strings.ReplaceAll(m[1], `\`, `/`)
		if _, err := os.Stat(steamDir); err != nil {
			// TODO: replace this hack with a generic case-insensitive path resolution
			steamDir = strings.ReplaceAll(steamDir, `steam/`, `Steam/`)
		}
	}
	if _, err := os.Stat(steamDir); err != nil {
		vm.Logs.Printf("readLineGamePath: Steam directory `%s` doesn't exist: %s\n", steamDir, err)
		return
	}

	gameDir := filepath.Join(steamDir, game.DirName)
	if _, err := os.Stat(gameDir); err != nil {
		vm.Logs.Printf("readLineGamePath: Game directory `%s` doesn't exist: %s\n", gameDir, err)
		return
	}

	vm.app.VoiceModGame(ts, game, gameDir)
}

func (vm *voiceMod) execStatus() error {
	return vm.Exec(X{"status"})
}

func (vm *voiceMod) readStatusTable(conn *telnet.Conn) {
	ts := time.Now()
	addr := ""
	if p := vm.statusServer.Load(); p != nil {
		addr = *p
	}
	var bots data.SliceSet[steam.Profile]
	var hums data.SliceSet[steam.Profile]
	for {
		ln, err := vm.Conn.ReadString('\n')
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
			bots = bots.Add(steam.Profile{Name: nm})
		} else {
			id, _ := steam.ParseID(id)
			p, _ := steam.FetchProfile(vm.app.Store(), id, nm)
			hums = hums.Add(p)
		}
	}

	vm.app.VoiceModPresence(ts, addr, hums, bots)
}

func (vm *voiceMod) readLineStatusServer(addr string) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}
	addr = host + ":" + port
	vm.statusServer.Store(&addr)
}

func (vm *voiceMod) readLineConnect(name, status string) {
	switch strings.ToLower(status) {
	case "connected":
		vm.execStatus()
	case "disconnected":
		vm.execStatus()
	case "not connected to server":
		vm.app.VoiceModServerDisconnected()
	}
}

func (vm *voiceMod) hostInGame(state appstate.AppState) (string, bool) {
	for _, p := range state.Presence.Humans.Slice() {
		if state.Hosts[p.Username] {
			return p.Username, true
		}
	}
	return "", false
}

func (vm *voiceMod) ignoreChat(state appstate.AppState, name string) (reason string) {
	pr := state.Presence

	if name == pr.Username {
		return ""
	}

	if state.ExcludeUsernames[name] || state.ExcludeUsernames["*"] {
		return "excluded"
	}

	if host, ok := vm.hostInGame(state); ok {
		return "host " + host + " is in game"
	}

	if !state.IncludeUsernames[name] && !state.IncludeUsernames["*"] {
		return "not included"
	}

	if !vm.app.Limiter(name).Allow() {
		return "rate limited"
	}

	return ""
}

func (vm *voiceMod) readLineChat(name, msg string) {
	state := vm.app.State()

	if r := vm.ignoreChat(state, name); r != "" {
		vm.Logs.Printf("readLineChat: ignored: `%s: %s`: %s\n", name, msg, r)
		return
	}

	au, err := sound.SoundOrTTS(vm.app.TTS(name), state.Config, name, msg)
	if err != nil {
		vm.Logs.Printf("voiceMod.readLine: username=`%s`, message=`%s`: %s\n", name, msg, err)
		return
	}

	select {
	case vm.Q <- au:
	case <-vm.Q:
		select {
		case vm.Q <- au:
		default:
			return
		}
	default:
	}
}

func (vm *voiceMod) readLine(line string) error {
	line = strings.TrimSpace(line)

	msg := strings.ToLower(line)
	if strings.Contains(msg, passwordRequiredMsg) || strings.Contains(msg, passwordIncorrectMsg) {
		return ErrPassword
	}

	if strings.ReplaceAll(line, " ", "") == StopWord {
		select {
		case vm.stop <- struct{}{}:
		default:
		}
		return nil
	}

	if line == StatusTableBegin {
		vm.readStatusTable(vm.Conn)
		return nil
	}

	if ln := StatusServerPat.FindStringSubmatch(line); len(ln) == 2 {
		vm.readLineStatusServer(ln[1])
		return nil
	}

	if ln := ChatPat.FindStringSubmatch(line); len(ln) == 3 {
		vm.readLineChat(ln[1], ln[2])
		return nil
	}

	if ln := CvarPat.FindStringSubmatch(line); len(ln) == 3 {
		vm.readLineCvar(ln[1], ln[2])
		return nil
	}

	if ln := ConnectPat.FindStringSubmatch(line); len(ln) == 3 {
		vm.readLineConnect(ln[1], ln[2])
		return nil
	}

	if ln := GamePathPat.FindStringSubmatch(line); len(ln) == 3 {
		vm.readLineGamePath(ln[1], ln[2])
		return nil
	}

	return nil
}

func (vm *voiceMod) pollHost(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			vm.execStatus()
		}
	}
}

func (vm *voiceMod) commandLoop(ctx context.Context) {
	if err := vm.Exec(X{"bind", "backspace", `echo ` + StopWord}); err != nil {
		vm.Logs.Println(err)
	}
	if err := vm.Exec(X{"path"}); err != nil {
		vm.Logs.Println(err)
	}
	if err := vm.execStatus(); err != nil {
		vm.Logs.Println(err)
	}

	vm.pollHost(ctx)
}

func (vm *voiceMod) Loop(ctx context.Context) error {
	go vm.playLoop(ctx)
	go vm.commandLoop(ctx)

	for {
		ln, err := vm.Conn.ReadString('\n')
		if err != nil {
			return err
		}
		if err := vm.readLine(ln); err != nil {
			return err
		}
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
			interval = rng.Range(maxInterval/2, maxInterval)
		}
		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return r, ctx.Err()
		}
	}
}

func initNetcon(app App) (c *telnet.Conn, err error) {
	defer func() {
		if err != nil && c != nil {
			c.Close()
		}
	}()

	nc := app.State().Netcon
	addr := nc.Addr()
	c, err = telnet.Dial("tcp", addr)

	if nc.Password != "" && err == nil {
		_, err = fmt.Fprintf(c, "PASS %s\r\n", nc.Password)
	}

	status := "connected"
	if err != nil {
		status = err.Error()
	}

	log := app.Logs().Debug
	if c != nil {
		log = app.Logs().Info
	}
	log("netcon",
		slog.String("addr", addr),
		slog.String("status", status),
		slog.Bool("password", nc.Password != ""),
	)

	return c, err
}

func dialTnet(ctx context.Context, app App) (_ *voiceMod, _ context.Context, cancel func(), _ error) {
	tc, err := retryForever(ctx, 5*time.Second, func() (*telnet.Conn, error) { return initNetcon(app) })
	if err != nil {
		return nil, ctx, nil, err
	}

	ctx, cancel = context.WithCancel(ctx)
	vm := &voiceMod{
		Q:    make(chan *audio.Audio, 1<<10),
		stop: make(chan struct{}, 1),
		Conn: tc,
		app:  app,
	}
	return vm, ctx, cancel, nil
}

func runVM(ctx context.Context, app App) (retErr error) {
	// reset any stale data. it will be re-initialized by readLineGamePath and readLineName
	defer func() { app.VoiceModStopped(retErr) }()

	vm, ctx, cancel, err := dialTnet(ctx, app)
	if err != nil {
		return err
	}
	defer cancel()

	return vm.Loop(ctx)
}

func Run(app App) {
	ctx := context.Background()
	for {
		err := runVM(ctx, app)
		switch {
		case errors.Is(err, ErrPassword):
			time.Sleep(30 * time.Second)
		default:
			// if the server is broken (e.g. broken pipe on write), but we can connect immediately
			// we end up just burning CPU, so always wait a little before restarting
			time.Sleep(5 * time.Second)
		}
	}
}

func quote(s string) string {
	for _, r := range s {
		if unicode.IsSpace(r) || !unicode.IsPrint(r) {
			return strconv.QuoteToASCII(s)
		}
	}
	return s
}
