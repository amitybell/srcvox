package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cockroachdb/pebble"
	"time"
)

var (
	ErrNilDB = errors.New("Use of uninitialized DB")
)

type DB struct {
	pb *pebble.DB
}

func (db *DB) Close() error {
	if db == nil {
		return fmt.Errorf("DB.Close: %w", ErrNilDB)
	}

	return db.pb.Close()
}

func (db *DB) Put(k string, v any) error {
	if db == nil {
		return fmt.Errorf("DB.Put(%s): %w", k, ErrNilDB)
	}

	s, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("DB.Put(%s): %w", k, err)
	}
	if err = db.pb.Set([]byte(k), s, pebble.Sync); err != nil {
		return fmt.Errorf("DB.Put(%s): %w", k, err)
	}
	return nil
}

func (db *DB) Get(k string, outPtr any) error {
	if db == nil {
		return fmt.Errorf("DB.Get(%s): %w", k, ErrNilDB)
	}

	s, c, err := db.pb.Get([]byte(k))
	if err != nil {
		return fmt.Errorf("DB.Get(%s): %w", k, err)
	}
	defer c.Close()
	return json.Unmarshal(s, outPtr)
}

func Put[T any](db *DB, k string, v T) error {
	return db.Put(k, v)
}

func Get[T any](db *DB, k string) (T, error) {
	var v T
	err := db.Get(k, &v)
	return v, err
}

type CacheEntry[T any] struct {
	Ts time.Time
	V  T
}

func Cache[T any](db *DB, ttl time.Duration, k string, new func() (T, error)) (T, error) {
	if ent, err := Get[CacheEntry[T]](db, k); err == nil && time.Since(ent.Ts) <= ttl {
		return ent.V, nil
	}

	v, err := new()
	if err != nil {
		return v, fmt.Errorf("Cache: %w", err)
	}

	// it's just a cache; it's fine if it fails
	_ = Put(db, k, CacheEntry[T]{Ts: time.Now(), V: v})

	return v, nil
}

func OpenDB(dir string) (*DB, error) {
	pb, err := pebble.Open(dir, nil)
	if err != nil {
		return nil, fmt.Errorf("OpenDB(%s): %w", dir, err)
	}
	return &DB{pb: pb}, nil
}
