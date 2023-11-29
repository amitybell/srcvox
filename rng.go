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

func randIntRange(n, m int) int {
	if m <= n {
		return m
	}
	return n + randIntn(m-n)
}

func randElem[T any](s []T) T {
	return s[randIntn(len(s))]
}
