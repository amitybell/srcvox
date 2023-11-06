package main

import (
	"context"
	"fmt"
	"github.com/amitybell/memio"
	"github.com/amitybell/piper"
	"github.com/amitybell/piper-asset"
	"github.com/amitybell/piper-voice-alan"
	"github.com/amitybell/piper-voice-jenny"
	"github.com/amitybell/srcvox/files"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"net/http"
	"os"
	"strconv"
	"sync"
)

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

	ctx  context.Context
	ttsl []*piper.TTS

	mu     sync.Mutex
	_state AppState
	ttsm   map[string]*piper.TTS
}

func NewApp(paths *Paths) *App {
	app := &App{
		Paths: paths,

		_state: DefaultAppState,
		ttsm:   map[string]*piper.TTS{},
	}
	app.API = &API{app: app}
	return app
}

func (a *App) Close() error {
	if a.DB != nil {
		a.DB.Close()
	}
	return nil
}

func (a *App) TTS(key string) *piper.TTS {
	a.mu.Lock()
	defer a.mu.Unlock()

	if tts, ok := a.ttsm[key]; ok {
		return tts
	}

	tts := a.ttsl[len(a.ttsm)%len(a.ttsl)]
	a.ttsm[key] = tts
	return tts
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx

	db, err := OpenDB(a.Paths.DBDir)
	if err != nil {
		a.FataError(err)
	}
	a.DB = db

	cfg, err := a.startup()
	if err != nil {
		a.FataError(err)
		return
	}
	a.Update(AppState{Config: cfg})

	go tnet(a)
}

func (a *App) startup() (Config, error) {
	cfg, err := readConfig(a.Paths.ConfigFn)
	if err != nil && !os.IsExist(err) {
		return cfg, err
	}

	for _, asset := range []asset.Asset{jenny.Asset, alan.Asset} {
		tts, err := piper.New("", asset)
		if err != nil {
			return cfg, err
		}
		a.ttsl = append(a.ttsl, tts)
	}

	return cfg, nil
}

func (a *App) FataError(err error) {
	a.UpdateState(func(s AppState) AppState {
		e := AppError{Fatal: true, Message: err.Error()}
		if prev := s.Error.Message; prev != "" {
			e.Message += "\n" + prev
		}
		s.Error = e
		return s
	})
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/app.sound", "/app.synthesize":
		a.serveSound(w, r)
	case "/app.gameicon":
		a.serveGameIcon(w, r)
	case "/favicon.ico":
		a.serveGameIcon(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (a *App) serveFavIcon(w http.ResponseWriter, r *http.Request) {
	w.Write(files.EmblemPNG)
}

func (a *App) serveGameIcon(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	mime, s, err := ReadGameIcon(id)
	if err != nil {
		w.Header().Set("Content-Type", mime)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Write(s)
}

func (a *App) serveSound(w http.ResponseWriter, r *http.Request) {
	state := a.State()
	pr := state.Presence
	au, err := SoundOrTTS(a.TTS(pr.Username), pr.Username, r.URL.Query().Get("text"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f := memio.NewFile(nil)
	if _, err := au.Encode(state, f, DefaultVoiceFormat); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(f.Bytes())
}

func (a *App) Emit(name string, data any) {
	runtime.EventsEmit(a.ctx, name, data)
}

func (a *App) State() AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a._state
}

func (a *App) UpdateState(f func(p AppState) AppState) {
	a.mu.Lock()
	defer a.mu.Unlock()

	s, events := a._state.Merge(f(a._state))
	a._state = s

	for _, name := range events {
		a.Emit(name, nil)
	}
}

func (a *App) Update(props AppState) {
	a.UpdateState(func(s AppState) AppState { return props })
}
