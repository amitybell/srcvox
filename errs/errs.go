package errs

import (
	"fmt"
	"runtime/debug"
)

func Recover(err *error) {
	e := recover()
	if e == nil {
		return
	}
	*err = fmt.Errorf("PANIC: %v\n%s\n", e, debug.Stack())
}
