package main

import (
	"errors"
)

var (
	ErrEmptyMessage   = errors.New("Empty message")
	ErrNotImplemented = errors.New("N/I")
)
