package main

import (
	"log"
	"os"
)

var (
	Logs = log.New(os.Stderr, "srcvox: ", log.Lshortfile|log.Ltime)
)

type LogWriter struct{ F *os.File }

func (w *LogWriter) Write(p []byte) (int, error) {
	if w.F != nil {
		// it's a best-effort attempt, it doesn't matter if it fails
		w.F.Write(p)
		// we can't guarantee the file will be closed, so sync just-in-case
		w.F.Sync()
	}
	return os.Stderr.Write(p)
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
