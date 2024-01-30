package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/amitybell/memio"
	"github.com/amitybell/piper"
	asset "github.com/amitybell/piper-asset"
	alan "github.com/amitybell/piper-voice-alan"
	jenny "github.com/amitybell/piper-voice-jenny"
	"github.com/amitybell/srcvox/appstate"
	"github.com/amitybell/srcvox/config"
	"github.com/amitybell/srcvox/data"
	"github.com/amitybell/srcvox/demo"
	"github.com/amitybell/srcvox/errs"
	"github.com/amitybell/srcvox/files"
	"github.com/amitybell/srcvox/logs"
	"github.com/amitybell/srcvox/sound"
	"github.com/amitybell/srcvox/steam"
	"github.com/amitybell/srcvox/store"
	"github.com/amitybell/srcvox/translate"
	"github.com/amitybell/srcvox/voicemod"
	"github.com/amitybell/srcvox/watch"
	"github.com/wailsapp/wails/v2/pkg/application"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/time/rate"
)

const (
	winConfKey = "/windows-conf"
)

type winConf struct {
	W         int
	H         int
	X         int
	Y         int
	Maximized bool
	Screen    struct {
		W int
		H int
	}
}

type App struct {
	API   *API
	DB    *store.DB
	Paths *config.Paths

	initErr []error

	listener net.Listener

	ctx  context.Context
	ttsl []*piper.TTS

	wapp *application.Application

	serveMux http.ServeMux

	state struct {
		p atomic.Pointer[appstate.AppState]
		q chan appstate.Reducer
	}

	tmr struct {
		reloadConfig *time.Timer
	}

	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	ttsm     map[string]*piper.TTS
}

func newStoppedTimer() *time.Timer {
	t := time.NewTimer(time.Hour)
	t.Stop()
	return t
}

func NewApp(paths *config.Paths) *App {
	app := &App{
		Paths:    paths,
		ttsm:     map[string]*piper.TTS{},
		limiters: map[string]*rate.Limiter{},
	}
	app.tmr.reloadConfig = newStoppedTimer()
	app.API = &API{app: app}

	app.state.p.Store(&appstate.AppState{})
	app.state.q = make(chan appstate.Reducer, 1<<10)
	go app.reduceLoop()

	var err error
	app.listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		app.initErr = append(app.initErr, fmt.Errorf("Cannot start local server: %w", err))
	}

	width, height := 0, 0
	if demo.Enabled {
		mul := 2
		width = 600 * mul
		height = 500 * mul
	}

	app.wapp = application.NewWithOptions(&options.App{
		Title:            "SrcVox",
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		Linux: &linux.Options{
			WebviewGpuPolicy: linux.WebviewGpuPolicyNever,
			Icon:             files.EmblemPNG,
		},
		Windows: &windows.Options{
			WebviewUserDataPath:  paths.WebviewDataDir,
			WebviewGpuIsDisabled: true,
		},
		Width:         width,
		Height:        height,
		OnStartup:     app.onStartup,
		OnShutdown:    app.onShutdown,
		OnBeforeClose: app.onBeforeClose,
		OnDomReady:    app.onDomReady,
		Bind:          []any{app.API},
		ErrorFormatter: func(err error) any {
			switch err := err.(type) {
			case nil:
				return nil
			case *appstate.AppError:
				return err
			default:
				return &appstate.AppError{Message: err.Error()}
			}
		},
		AssetServer: &assetserver.Options{
			Assets:  assetsFS,
			Handler: app,
		},
		LogLevel: logger.ERROR,
	})

	return app
}

func (app *App) Run() error {
	return app.wapp.Run()
}

func (app *App) Close() error {
	if app.DB != nil {
		app.DB.Close()
	}
	return nil
}

func (app *App) TTS(key string) *piper.TTS {
	app.mu.Lock()
	defer app.mu.Unlock()

	if tts, ok := app.ttsm[key]; ok {
		return tts
	}

	tts := app.ttsl[len(app.ttsm)%len(app.ttsl)]
	app.ttsm[key] = tts
	return tts
}

func (app *App) screenSize(ctx context.Context) (int, int) {
	screens, _ := runtime.ScreenGetAll(ctx)
	for _, s := range screens {
		if s.IsCurrent {
			return s.Width, s.Height
		}
	}
	return 0, 0
}

func (app *App) onShutdown(ctx context.Context) {
}

func (app *App) onBeforeClose(ctx context.Context) bool {
	app.saveWinConf(ctx)

	return false
}

func (app *App) onDomReady(ctx context.Context) {
}

func (app *App) winConfKey(ctx context.Context) string {
	w, h := app.screenSize(ctx)
	return fmt.Sprintf("%s/%dx%d", winConfKey, w, h)
}

func (app *App) saveWinConf(ctx context.Context) {
	if demo.Enabled {
		return
	}

	wc := winConf{}
	wc.Maximized = runtime.WindowIsMaximised(ctx)
	wc.Screen.W, wc.Screen.H = app.screenSize(ctx)
	if !wc.Maximized {
		wc.W, wc.H = runtime.WindowGetSize(ctx)
		wc.X, wc.Y = runtime.WindowGetPosition(ctx)
	}
	app.DB.Put(app.winConfKey(ctx), wc)
}

func (app *App) initWinConf(ctx context.Context) {
	if demo.Enabled {
		runtime.WindowSetSize(ctx, demo.Width, demo.Height)
		return
	}

	wc, _ := store.Get[winConf](app.DB, app.winConfKey(ctx))
	wc.W = max(wc.W, 64*16)
	wc.H = max(wc.H, 32*16)

	sW, sH := app.screenSize(ctx)
	if wc.X > 0 && wc.Y > 0 && wc.Screen.W == sW && wc.Screen.H == sH {
		runtime.WindowSetPosition(ctx, wc.X, wc.Y)
	}
	if wc.W > 0 && wc.H > 0 {
		runtime.WindowSetSize(ctx, wc.W, wc.H)
	}
	switch {
	case app.State().StartMinimized():
		runtime.WindowMinimise(ctx)
	case wc.Maximized:
		runtime.WindowMaximise(ctx)
	}
}

func (app *App) Limiter(name string) *rate.Limiter {
	app.mu.Lock()
	defer app.mu.Unlock()

	lim, ok := app.limiters[name]
	if !ok {
		lim = rate.NewLimiter(rate.Every(app.State().RateLimit.D), 1)
		app.limiters[name] = lim
	}
	return lim
}

func (app *App) initWatch() {
	watch.Path(app.Paths.ConfigDir)

	watch.Notify(func(ev watch.Event) {
		switch {
		case ev.Name == app.Paths.ConfigFn:
			app.tmr.reloadConfig.Reset(2 * time.Second)
		case filepath.Base(ev.Name) == "loginusers.vdf":
			app.initPresence()
		}
	})
}

func (app *App) initDB() error {
	var err error
	app.DB, err = store.OpenDB(app.Paths.DBDir, store.Options{Logger: Logs.Pebble(false)})
	return err
}

func (app *App) reloadConfig() {
	for range app.tmr.reloadConfig.C {
		cfg, err := config.Read(app.Paths.ConfigFn)
		if err != nil {
			Logs.Println("reloadConfig:", err)
			continue
		}
		app.API.app.UpdateState(func(s appstate.AppState) appstate.AppState {
			s.Config = cfg
			Logs.Println("reloadConfig: ok")
			return s
		})
	}
}

func (app *App) initConfig() (config.Config, error) {
	go app.reloadConfig()

	cfg, err := config.Read(app.Paths.ConfigFn)
	if err != nil && !os.IsExist(err) {
		return cfg, err
	}
	app.Update(appstate.AppState{Config: cfg})
	return cfg, nil
}

func (app *App) initPiper(firstVoice string) error {
	voices := []asset.Asset{jenny.Asset, alan.Asset}
	if firstVoice == "alan" {
		voices = []asset.Asset{alan.Asset, jenny.Asset}
	}
	for _, asset := range voices {
		tts, err := piper.New("", asset)
		if err != nil {
			return err
		}
		app.ttsl = append(app.ttsl, tts)
	}
	return nil
}

func (app *App) initPresence() {
	usr, ok := steam.FindUser(app.DB, 0)
	if !ok {
		return
	}
	app.UpdateState(func(s appstate.AppState) appstate.AppState {
		p := s.Presence
		p.InGame = demo.Enabled
		p.UserID = usr.ID
		p.Username = usr.Name
		p.AvatarURI = steam.AvatarURI(app.DB, usr.ID)
		if demo.Enabled {
			p.Username = demo.Username
			p.AvatarURI = data.URI("image/jpeg", files.DemoAvatar)
		}
		p.Clan, p.Name = translate.ClanName(usr.Name)
		p.Ts = time.Now()
		s.Presence = p
		return s
	})

}

func (app *App) onStartup(ctx context.Context) {
	app.ctx = ctx

	cfg, err := app.initConfig()
	if err != nil {
		app.FatalError(err)
		return
	}

	if err := app.initDB(); err != nil {
		app.FatalError(err)
		return
	}

	app.initWinConf(ctx)

	if len(app.initErr) != 0 {
		for _, err := range app.initErr {
			app.FatalError(err)
		}
		return
	}

	app.initPresence()

	if err := app.initPiper(cfg.FirstVoice); err != nil {
		app.FatalError(err)
		return
	}

	app.initWatch()
	app.initServer()

	go voicemod.Run(app)
}

func (app *App) Error(fatal bool, err error) {
	app.UpdateState(func(s appstate.AppState) appstate.AppState {
		e := appstate.AppError{
			Fatal:   fatal || s.Error.Fatal,
			Message: err.Error(),
		}
		if prev := s.Error.Message; prev != "" {
			e.Message += "\n" + prev
		}
		s.Error = e
		return s
	})
}

func (app *App) FatalError(err error) {
	app.Error(true, err)
}

func (app *App) serveFavicon(w http.ResponseWriter, r *http.Request) {
	w.Write(files.EmblemPNG)
}

type fileContent struct {
	f *os.File
}

func (app *App) serveContent(w http.ResponseWriter, r *http.Request, rd io.Reader) {
	rs, ok := rd.(io.ReadSeeker)
	if !ok {
		io.Copy(w, rd)
		return
	}

	name := ""
	mtime := time.Time{}
	if f, ok := rs.(interface {
		Stat() (os.FileInfo, error)
		Name() string
	}); ok {
		name = f.Name()
		if fi, err := f.Stat(); err == nil {
			mtime = fi.ModTime()
			w.Header().Set("ETag", fmt.Sprintf(`"%x.%x"`, fi.Size(), mtime.UnixNano()))
		}
	}
	http.ServeContent(w, r, name, mtime, rs)
}

func (app *App) serveMapImage(w http.ResponseWriter, r *http.Request) {
	qry := r.URL.Query()
	id := qry.Get("id")
	nm := qry.Get("map")

	g, ok := steam.GamesMapString[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	if f, err := g.OpenMapImage(nm); err == nil {
		defer f.Close()
		app.serveContent(w, r, f)
	}

	// we want to display _something_, otherwise the server list entry looks out-of-place
	mime, s, err := steam.ReadGameImage(g.ID, steam.HeroImage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Write(s)
}

func (app *App) serveAvatar(w http.ResponseWriter, r *http.Request) {
	if demo.Enabled {
		w.Write(files.DemoAvatar)
		return
	}

	id, _ := steam.ParseID(r.URL.Query().Get("id"))
	f, err := steam.OpenAvatar(app.DB, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer f.Close()
	app.serveContent(w, r, f)
}

func (app *App) serveBgVideo(w http.ResponseWriter, r *http.Request) {
	g, ok := steam.GamesMapString[r.URL.Query().Get("id")]
	if !ok {
		http.NotFound(w, r)
		return
	}
	f, err := g.OpenBgVideo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer f.Close()
	app.serveContent(w, r, f)
}

func (app *App) serveSound(w http.ResponseWriter, r *http.Request) {
	state := app.State()
	pr := state.Presence
	au, err := sound.SoundOrTTS(app.TTS(pr.Username), state.Config, pr.Username, r.URL.Query().Get("text"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f := memio.NewFile(nil)
	if _, err := au.Encode(state.AudioDelay.D, state.AudioLimit.D, f, voicemod.DefaultVoiceFormat); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.serveContent(w, r, f)
}

func (app *App) initServeMux() {
	app.serveMux.HandleFunc("/app.bgvideo", app.serveBgVideo)
	app.serveMux.HandleFunc("/app.mapimage", app.serveMapImage)
	app.serveMux.HandleFunc("/app.sound", app.serveSound)
	app.serveMux.HandleFunc("/app.synthesize", app.serveSound)
	app.serveMux.HandleFunc("/app.avatar", app.serveAvatar)
	app.serveMux.HandleFunc("/favicon.ico", app.serveFavicon)
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		return
	}
	app.serveMux.ServeHTTP(w, r)
}

func (app *App) serverURL(path string, query url.Values) (string, bool) {
	lsn := app.listener
	if lsn == nil {
		return "", false
	}
	u := &url.URL{
		Scheme:   "http",
		Host:     lsn.Addr().String(),
		Path:     path,
		RawQuery: query.Encode(),
	}
	return u.String(), true
}

func (app *App) serve(lsn net.Listener) {
	err := http.Serve(app.listener, app)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		app.FatalError(err)
	}
}

func (app *App) initServer() {
	if app.listener == nil {
		return
	}

	app.initServeMux()
	go app.serve(app.listener)
}

func (app *App) Emit(name string, data any) {
	runtime.EventsEmit(app.ctx, name, data)
}

func (app *App) Store() *store.DB {
	return app.DB
}

func (app *App) State() appstate.AppState {
	return *app.state.p.Load()
}

func (app *App) reduce(reduce appstate.Reducer) {
	defer func() {
		var err error
		errs.Recover(&err)
		if err != nil {
			Logs.Error("App.reduce:", err)
		}
	}()

	oldState := *app.state.p.Load()
	newState := reduce(oldState)
	state, events := oldState.Merge(newState)
	app.state.p.Store(&state)

	for _, name := range events {
		app.Emit(name, nil)
	}
}

func (app *App) reduceLoop() {
	for f := range app.state.q {
		app.reduce(f)
	}
}

func (app *App) UpdateState(f appstate.Reducer) {
	app.state.q <- f
}

func (app *App) Update(props appstate.AppState) {
	app.UpdateState(func(s appstate.AppState) appstate.AppState { return props })
}

func (app *App) UpdateConfig(cfg config.Config) error {
	done := make(chan error)
	app.UpdateState(func(s appstate.AppState) appstate.AppState {
		var err error
		cfg, changed := s.Config.Merge(cfg)
		if changed {
			err = config.Write(app.Paths.ConfigFn, cfg)
		}
		s.Config = cfg
		done <- err
		return s
	})
	return <-done
}

func (app *App) Logs() *logs.Logger {
	return Logs
}

func (app *App) VoiceModServerDisconnected() {
	app.UpdateState(func(s appstate.AppState) appstate.AppState {
		s.Presence.Humans = s.Presence.Humans.Clear()
		s.Presence.Bots = s.Presence.Bots.Clear()
		return s
	})
}

func (app *App) VoiceModPresence(ts time.Time, server string, hums, bots data.SliceSet[steam.Profile]) {
	app.UpdateState(func(s appstate.AppState) appstate.AppState {
		if s.Presence.Server == server &&
			s.Presence.Bots.Equal(bots) &&
			s.Presence.Humans.Equal(hums) {
			return s
		}
		s.Presence.Bots = bots
		s.Presence.Humans = hums
		s.Presence.Server = server
		s.Presence.Ts = ts
		return s
	})

}

func (app *App) VoiceModGame(ts time.Time, game *steam.GameInfo, gameDir string) {
	app.UpdateState(func(s appstate.AppState) appstate.AppState {
		s.Presence.InGame = true
		s.Presence.GameID = game.ID
		s.Presence.GameIconURI = game.IconURI
		s.Presence.GameHeroURI = game.HeroURI
		s.Presence.GameDir = gameDir
		s.Presence.Ts = ts
		return s
	})
}

func (app *App) VoiceModStopped(err error) {
	app.UpdateState(func(s appstate.AppState) appstate.AppState {
		switch {
		case errors.Is(err, voicemod.ErrPassword):
			s.Presence.Error = "Cannot connect to Game: " + err.Error()
		default:
			s.Presence.Error = ""
		}
		s.Presence.InGame = demo.Enabled
		s.Presence.Humans = s.Presence.Humans.Clear()
		s.Presence.Bots = s.Presence.Bots.Clear()
		s.Presence.Server = ""
		return s
	})

}
