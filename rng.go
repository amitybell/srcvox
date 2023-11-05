package main

import (
	"lukechampine.com/frand"
)

func randBool() bool {
	return frand.Intn(1<<10)&1 == 1
}

func randInt() int {
	return int(frand.Intn(1<<32 - 1))
}

func randElem[T any](s []T) T {
	return s[frand.Intn(len(s))]
}
