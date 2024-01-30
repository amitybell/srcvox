package watch

import (
	"path/filepath"
	"sync"

	"github.com/amitybell/srcvox/errs"
	"github.com/amitybell/srcvox/logs"
	"github.com/fsnotify/fsnotify"
)

type Event struct {
	fsnotify.Event
}

var (
	Logs = logs.AppLogger()

	watch = struct {
		sync.Mutex
		fsw    *fsnotify.Watcher
		seen   map[string]bool
		notify []func(ev Event)
	}{
		seen: map[string]bool{},
	}
)

func Path(fn string) error {
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

func watchNotify(ev Event, f func(Event)) {
	defer func() {
		var err error
		errs.Recover(&err)
		if err != nil {
			Logs.Println("Watch.Notify PANIC:", err)
		}
	}()

	f(ev)
}

func watchNotifier(fsw *fsnotify.Watcher) {
	for e := range fsw.Events {
		ev := Event{Event: e}
		watch.Lock()
		notify := watch.notify
		watch.Unlock()

		for _, f := range notify {
			go watchNotify(ev, f)
		}
	}
}

func Notify(f func(ev Event)) {
	watch.Lock()
	defer watch.Unlock()

	watch.notify = append(watch.notify[:len(watch.notify):len(watch.notify)], f)
}
