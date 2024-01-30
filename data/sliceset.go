package data

import (
	"encoding/json"
	"unsafe"
)

var (
	_ json.Marshaler = SliceSet[any]{}
)

type SliceSet[T comparable] struct {
	p *T
	n int
}

func (s SliceSet[T]) Len() int {
	return s.n
}

func (s SliceSet[T]) Slice() []T {
	return unsafe.Slice(s.p, s.n)
}

func (s SliceSet[T]) append(v T) SliceSet[T] {
	a := append(s.Slice(), v)
	return SliceSet[T]{p: unsafe.SliceData(a), n: len(a)}
}

func (s SliceSet[T]) Index(v T) (int, bool) {
	for i, x := range s.Slice() {
		if v == x {
			return i, true
		}
	}
	return -1, false
}

func (s SliceSet[T]) Is(t SliceSet[T]) bool {
	return s == t
}

func (s SliceSet[T]) Equal(t SliceSet[T]) bool {
	p := s.Slice()
	q := t.Slice()
	if len(p) != len(q) {
		return false
	}
	for i, v := range p {
		if q[i] != v {
			return false
		}
	}
	return true
}

func (s SliceSet[T]) Has(v T) bool {
	_, ok := s.Index(v)
	return ok
}

func (s SliceSet[T]) Add(v T) SliceSet[T] {
	if s.Has(v) {
		return s
	}
	return s.append(v)
}

func (s SliceSet[T]) Clear() SliceSet[T] {
	return SliceSet[T]{}
}

func (s SliceSet[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Slice())
}
