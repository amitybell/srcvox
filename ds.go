package main

import (
	"unsafe"
	"encoding/json"
)

var (
	_ json.Marshaler= SliceSet[any]{}
)

type SliceSet[T comparable] struct{
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

func (s SliceSet[T]) Index(v T) (int,bool) {
	for i,x:=range s.Slice(){
		if v ==x {
			return i,true
		}
	}
	return -1,false
}


func (s SliceSet[T]) Has(v T) bool {
	_,ok:=s.Index(v)
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

