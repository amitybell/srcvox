package platform

import "os"

func SignalIsTerm(sig os.Signal) bool {
	for _, s := range TermSignals {
		if sig == s {
			return true
		}
	}
	return false
}
