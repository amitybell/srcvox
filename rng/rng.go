package rng

import (
	"lukechampine.com/frand"
)

func Intn(n int) int {
	return int(frand.Intn(n))
}

func Int() int {
	return Intn(1<<32 - 1)
}

func Bool() bool {
	return Int()&1 == 1
}

func Range[T ~int8 | ~int16 | ~int32 | ~int64 | ~int](n, m T) T {
	if m <= n {
		return m
	}
	return n + T(Intn(int(m-n)))
}

func Elem[T any](s []T) T {
	return s[Intn(len(s))]
}

func Shuffle[S ~[]E, E any](s S) S {
	s = append([]E(nil), s...)
	frand.Shuffle(len(s), func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})
	return s
}
