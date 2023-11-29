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
	"sync"
	"time"

	"github.com/amitybell/memio"
	"github.com/amitybell/piper"
	asset "github.com/amitybell/piper-asset"
	alan "github.com/amitybell/piper-voice-alan"
	jenny "github.com/amitybell/piper-voice-jenny"
	"github.com/amitybell/srcvox/files"
	"github.com/wailsapp/wails/v2/pkg/application"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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

type AppError struct {
	Fatal   bool   `json:"fatal"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func AppErr(pfx string, err error) *AppError {
	if err != nil {
		if pfx != "" {
			return &AppError{Message: fmt.Sprintf("%s: %s", pfx, err)}
		}
		return &AppError{Message: err.Error()}
	}
	return nil
}

type App struct {
	API   *API
	DB    *DB
	Paths *Paths

	initErr []error

	listener net.Listener

	ctx  context.Context
	ttsl []*piper.TTS

	wapp *application.Application

	serveMux http.ServeMux

	mu     sync.Mutex
	_state AppState
	ttsm   map[string]*piper.TTS
}

func NewApp(paths *Paths) *App {
	app := &App{
		Paths:  paths,
		_state: DefaultAppState,
		ttsm:   map[string]*piper.TTS{},
	}
	app.API = &API{app: app}

	var err error
	app.listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		app.initErr = append(app.initErr, fmt.Errorf("Cannot start local server: %w", err))
	}

	width, height := 0, 0
	if Env.Demo {
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
			case *AppError:
				return err
			default:
				return &AppError{Message: err.Error()}
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
	if Env.Demo {
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
	if Env.Demo {
		runtime.WindowSetSize(ctx, DemoWidth, DemoHeight)
		return
	}

	wc, _ := Get[winConf](app.DB, app.winConfKey(ctx))
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
	case Env.Minimized:
		runtime.WindowMinimise(ctx)
	case wc.Maximized:
		runtime.WindowMaximise(ctx)
	}
}

func (app *App) initDB() error {
	var err error
	app.DB, err = OpenDB(app.Paths.DBDir)
	return err
}

func (app *App) initConfig() (Config, error) {
	cfg, err := readConfig(app.Paths.ConfigFn)
	if err != nil && !os.IsExist(err) {
		return cfg, err
	}
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
	usr, _, ok := findSteamUser()
	if !ok {
		return
	}
	app.UpdateState(func(s AppState) AppState {
		p := s.Presence
		p.InGame = Env.Demo
		p.UserID = usr.ID
		p.Username = usr.Name
		p.Clan, p.Name = ClanName(usr.Name)
		p.AvatarURL, _ = app.serverURL("/app.avatar", nil)
		s.Presence = p
		return s
	})

}

func (app *App) onStartup(ctx context.Context) {
	app.ctx = ctx

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

	cfg, err := app.initConfig()
	if err != nil {
		app.FatalError(err)
		return
	}
	app.Update(AppState{Config: cfg})

	if err := app.initPiper(cfg.FirstVoice); err != nil {
		app.FatalError(err)
		return
	}

	app.initServer()

	go tnet(app)
}

func (app *App) Error(fatal bool, err error) {
	app.UpdateState(func(s AppState) AppState {
		e := AppError{
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

	g, ok := GamesMapString[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	if f, err := g.OpenMapImage(nm); err == nil {
		defer f.Close()
		app.serveContent(w, r, f)
	}

	// we want to display _something_, otherwise the server list entry looks out-of-place
	mime, s, err := ReadGameImage(g.ID, HeroImage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Write(s)
}

func (app *App) serveAvatar(w http.ResponseWriter, r *http.Request) {
	if Env.Demo {
		w.Write(files.DemoAvatar)
		return
	}

	f, err := openUserAvatar()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer f.Close()
	app.serveContent(w, r, f)
}

func (app *App) serveBgVideo(w http.ResponseWriter, r *http.Request) {
	g, ok := GamesMapString[r.URL.Query().Get("id")]
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
	au, err := SoundOrTTS(app.TTS(pr.Username), state, pr.Username, r.URL.Query().Get("text"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f := memio.NewFile(nil)
	if _, err := au.Encode(state, f, DefaultVoiceFormat); err != nil {
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

func (app *App) State() AppState {
	app.mu.Lock()
	defer app.mu.Unlock()

	return app._state
}

func (app *App) UpdateState(f func(p AppState) AppState) {
	app.mu.Lock()
	defer app.mu.Unlock()

	s, events := app._state.Merge(f(app._state))
	app._state = s

	for _, name := range events {
		app.Emit(name, nil)
	}
}

func (app *App) Update(props AppState) {
	app.UpdateState(func(s AppState) AppState { return props })
}
