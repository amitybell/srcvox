package main

import (
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrNilDB = errors.New("Use of uninitialized DB")
	ErrStale = errors.New("Stale")
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

	s, err := msgpack.Marshal(v)
	if err != nil {
		return fmt.Errorf("DB.Put(%s): Marshal: %w", k, err)
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

	if err := msgpack.Unmarshal(s, outPtr); err != nil {
		return fmt.Errorf("DB.Get(%s): Unmarshal: %w", k, err)
	}
	return nil
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
	Ts  time.Time
	Ver int
	V   T
}

func (ent CacheEntry[T]) CheckTTL(ttl time.Duration, ver int) bool {
	return ent.Ver == ver && (ttl < 0 || time.Since(ent.Ts) <= ttl)
}

func (ent CacheEntry[T]) CheckMtime(mtime time.Time, ver int) bool {
	return ent.Ver == ver && ent.Ts.Equal(mtime)
}

func CacheTTL[T any](db *DB, ttl time.Duration, k string, ver int, new func() (T, error)) (T, error) {
	ent, entErr := Get[CacheEntry[T]](db, k)
	if entErr == nil && ent.CheckTTL(ttl, ver) {
		return ent.V, nil
	}

	v, err := new()
	if err != nil {
		if entErr == nil {
			return ent.V, fmt.Errorf("CacheTTL: %w", errors.Join(err, ErrStale))
		}
		return v, fmt.Errorf("CacheTTL: %w", err)
	}

	// it's just a cache; it's fine if it fails
	_ = Put(db, k, CacheEntry[T]{Ts: time.Now().UTC(), V: v, Ver: ver})

	return v, nil
}

func CacheStat[T any](db *DB, f interface{ Stat() (fs.FileInfo, error) }, k string, ver int, new func() (T, error)) (T, error) {
	var mtime time.Time
	if fi, err := f.Stat(); err == nil {
		mtime = fi.ModTime().UTC()
	}

	ent, entErr := Get[CacheEntry[T]](db, k)
	if entErr == nil && ent.CheckMtime(mtime, ver) {
		return ent.V, nil
	}

	v, err := new()
	if err != nil {
		if entErr == nil {
			return ent.V, fmt.Errorf("CacheStat: %w", errors.Join(err, ErrStale))
		}
		return v, fmt.Errorf("CacheStat: %w", err)
	}

	// it's just a cache; it's fine if it fails
	_ = Put(db, k, CacheEntry[T]{Ts: mtime, V: v, Ver: ver})

	return v, nil
}

func CacheMtime[T any](db *DB, mtime time.Time, k string, ver int, new func() (T, error)) (T, error) {
	ent, entErr := Get[CacheEntry[T]](db, k)
	if entErr == nil && ent.CheckMtime(mtime, ver) {
		return ent.V, nil
	}

	v, err := new()
	if err != nil {
		if entErr == nil {
			return ent.V, fmt.Errorf("CacheMtime: %w", errors.Join(err, ErrStale))
		}
		return v, fmt.Errorf("CacheMtime: %w", err)
	}

	// it's just a cache; it's fine if it fails
	_ = Put(db, k, CacheEntry[T]{Ts: mtime, V: v, Ver: ver})

	return v, nil
}

func OpenDB(dir string) (*DB, error) {
	opts := &pebble.Options{
		Logger: &pebbleLogger{},
	}
	pb, err := pebble.Open(dir, opts)
	if err != nil {
		return nil, fmt.Errorf("OpenDB(%s): %w", dir, err)
	}
	return &DB{pb: pb}, nil
}
