package main

import (
	"log"
	"os"
	"path/filepath"
)

type logWriter struct{ F *os.File }

func (w *logWriter) Write(p []byte) (int, error) {
	if w.F != nil {
		// it's a best-effort attempt, it doesn't matter if it fails
		w.F.Write(p)
		// we can't guarantee the file will be closed, so sync just-in-case
		w.F.Sync()
	}
	return os.Stderr.Write(p)
}

var Logs, logsFile = func() (*log.Logger, *os.File) {
	fn := filepath.Join(DataDir, "logs.txt")
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Cannot open %s: %s", fn, err)
	}
	l := log.New(&logWriter{F: f}, "srcvox: ", log.Lshortfile|log.Ltime)
	return l, f
}()
