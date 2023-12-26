package main

import (
	"lukechampine.com/frand"
)

func randIntn(n int) int {
	return int(frand.Intn(n))
}

func randInt() int {
	return randIntn(1<<32 - 1)
}

func randBool() bool {
	return randInt()&1 == 1
}

func randRange[T ~int8 | ~int16 | ~int32 | ~int64 | ~int](n, m T) T {
	if m <= n {
		return m
	}
	return n + T(randIntn(int(m-n)))
}

func randElem[T any](s []T) T {
	return s[randIntn(len(s))]
}
