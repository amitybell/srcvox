package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/gofrs/uuid/v5"
)

var (
	osStderr = os.Stderr

	Logs = func() *Logger {
		lg := NewLogger(paths.LogsFn)
		if lg.F != nil {
			os.Stderr = lg.F
		}
		log.SetOutput(lg)
		return lg
	}()

	_ pebble.Logger = (*pebbleLogger)(nil)
)

type pebbleLogger struct {
	LogInfo bool
}

func (p *pebbleLogger) Infof(format string, args ...any) {
	if !p.LogInfo {
		return
	}
	Logs.Record(1, slog.LevelInfo, fmt.Sprintf("pebble.INFO: "+format, args...))
}

func (p *pebbleLogger) Errorf(format string, args ...any) {
	Logs.Record(1, slog.LevelError, fmt.Sprintf("pebble.ERROR: "+format, args...))
}

func (p *pebbleLogger) Fatalf(format string, args ...any) {
	Logs.Record(1, slog.LevelError, fmt.Sprintf("pebble.FATAL: "+format, args...))
}

type LogHandler struct {
	h slog.Handler
}

func (l *LogHandler) Handle(ctx context.Context, r slog.Record) error {
	if !l.Enabled(ctx, r.Level) {
		return nil
	}

	if id, err := uuid.NewV7(); err == nil {
		r.AddAttrs(slog.String("logID", id.String()))
	}
	if r.PC != 0 {
		f := runtime.FuncForPC(r.PC)
		if f != nil {
			fn, ln := f.FileLine(r.PC)
			r.AddAttrs(slog.String("src", fmt.Sprintf("%s:%d", fn, ln)))
		}
	}
	return l.h.Handle(ctx, r)
}

func (l *LogHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return l.h.Enabled(ctx, lvl)
}

func (l *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return l.h.WithAttrs(attrs)
}

func (l *LogHandler) WithGroup(name string) slog.Handler {
	return l.h.WithGroup(name)
}

type Logger struct {
	F *os.File
	*slog.Logger

	mu sync.Mutex
}

func (l *Logger) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.F != nil {
		// it's a best-effort attempt, it doesn't matter if it fails
		l.F.Write(p)
		// we can't guarantee the file will be closed, so sync just-in-case
		l.F.Sync()
	}
	return osStderr.Write(p)
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if f := l.F; f != nil {
		l.F = nil
		return f.Close()
	}
	return nil
}

func (l *Logger) Record(skip int, lvl slog.Level, msg string, attrs ...slog.Attr) error {
	if !l.Enabled(context.Background(), lvl) {
		return nil
	}

	skip += 2 // Callers + Record
	var pcs [1]uintptr
	runtime.Callers(skip, pcs[:])
	r := slog.NewRecord(time.Now(), lvl, msg, pcs[0])
	r.AddAttrs(attrs...)
	return l.Handler().Handle(context.Background(), r)
}

func (l *Logger) Printf(f string, a ...any) {
	l.Record(1, slog.LevelInfo, fmt.Sprintf(f, a...))
}

func (l *Logger) Println(a ...any) {
	l.Record(1, slog.LevelInfo, fmt.Sprintln(a...))
}

func (l *Logger) API(v APILog) {
	lvl := LogLevel(v.Level)
	var attrs []slog.Attr
	if len(v.Trace) != 0 {
		trace := v.Trace[len(v.Trace)-1]
		if lvl == slog.LevelDebug {
			trace = strings.Join(v.Trace, "\n")
		}
		attrs = append(attrs, slog.String("trace", trace))
	}
	l.Record(1, lvl, v.Message, attrs...)
}

func NewLogger(fn string) *Logger {
	cfg, _ := readConfig(paths.ConfigFn)
	lh := &LogHandler{}
	lg := &Logger{}
	lh.h = slog.NewJSONHandler(lg, &slog.HandlerOptions{Level: LogLevel(cfg.LogLevel)})
	lg.F, _ = os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	lg.Logger = slog.New(lh)
	return lg
}

func LogLevel(lvl string) slog.Level {
	switch strings.ToLower(lvl) {
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	case "debug":
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}
