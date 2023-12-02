package main

import (
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type WatchEvent struct {
	fsnotify.Event
}

var (
	watch = struct {
		sync.Mutex
		fsw    *fsnotify.Watcher
		seen   map[string]bool
		notify []func(ev WatchEvent)
	}{
		seen: map[string]bool{},
	}
)

func Watch(fn string) error {
	fn = filepath.Clean(fn)

	watch.Lock()
	defer watch.Unlock()

	if watch.seen[fn] {
		return nil
	}
	watch.seen[fn] = true

	if watch.fsw == nil {
		fsw, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		watch.fsw = fsw
		go watchNotifier(fsw)
	}

	return watch.fsw.Add(fn)
}

func watchNotify(ev WatchEvent, f func(WatchEvent)) {
	defer func() {
		var err error
		recoverPanic(&err)
		if err != nil {
			Logs.Println("Watch.Notify PANIC:", err)
		}
	}()

	f(ev)
}

func watchNotifier(fsw *fsnotify.Watcher) {
	for e := range fsw.Events {
		ev := WatchEvent{Event: e}
		watch.Lock()
		notify := watch.notify
		watch.Unlock()

		for _, f := range notify {
			go watchNotify(ev, f)
		}
	}
}

func WatchNotify(f func(ev WatchEvent)) {
	watch.Lock()
	defer watch.Unlock()

	watch.notify = append(watch.notify[:len(watch.notify):len(watch.notify)], f)
}
