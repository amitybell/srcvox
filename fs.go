package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeFile(fn string, hdl func(*os.File) error) error {
	// ensure fn is absolute, to avoid any race conditions with chdir
	dir, err := filepath.Abs(filepath.Dir(fn))
	if err != nil {
		return fmt.Errorf("writeFile: dir(%s): %w", fn, err)
	}

	tmp, err := os.CreateTemp(dir, ".srcvox.")
	if err != nil {
		return fmt.Errorf("writeFile: create temp: %w", err)
	}
	// we handle close(flush) below. this is to ensure we never leave files lying around
	defer os.Remove(tmp.Name())

	if err := hdl(tmp); err != nil {
		return fmt.Errorf("writeFile handle: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("writeFile close: %w", err)
	}

	if err := os.Rename(tmp.Name(), fn); err != nil {
		return fmt.Errorf("writeFile rename: %w", err)
	}
	return nil
}

func writeBytes(fn string, v []byte) error {
	return writeFile(fn, func(f *os.File) error {
		_, err := f.Write(v)
		if err != nil {
			return fmt.Errorf("writeBytes: %w", err)
		}
		return nil
	})
}

func writeJSON(fn string, indent string, v any) error {
	return writeFile(fn, func(f *os.File) error {
		enc := json.NewEncoder(f)
		if indent != "" {
			enc.SetIndent("", indent)
		}
		err := enc.Encode(f)
		if err != nil {
			return fmt.Errorf("writeJSON: encode: %w", err)
		}
		return nil
	})
}
