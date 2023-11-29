package main

import (
	"log"
	"os"

	"github.com/cockroachdb/pebble"
)

var (
	osStderr = os.Stderr
	Logs     = log.New(osStderr, "srcvox: ", log.Lshortfile|log.Ltime)

	_ pebble.Logger = (*pebbleLogger)(nil)
)

type LogWriter struct{ F *os.File }

func (w *LogWriter) Write(p []byte) (int, error) {
	if w.F != nil {
		// it's a best-effort attempt, it doesn't matter if it fails
		w.F.Write(p)
		// we can't guarantee the file will be closed, so sync just-in-case
		w.F.Sync()
	}
	return osStderr.Write(p)
}

func (w *LogWriter) Close() error {
	if w.F != nil {
		return w.F.Close()
	}
	return nil
}

func NewLogWriter(fn string) *LogWriter {
	f, _ := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	return &LogWriter{F: f}
}

type pebbleLogger struct {
	LogInfo bool
}

func (p *pebbleLogger) Infof(format string, args ...any) {
	if !p.LogInfo {
		return
	}
	Logs.Printf("pebble.INFO: "+format, args...)
}

func (p *pebbleLogger) Errorf(format string, args ...any) {
	Logs.Printf("pebble.ERROR: "+format, args...)
}

func (p *pebbleLogger) Fatalf(format string, args ...any) {
	Logs.Printf("pebble.FATAL: "+format, args...)
}
