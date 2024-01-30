package steam

import (
	"fmt"
	"os"

	"github.com/andygrunwald/vdf"
	"github.com/mitchellh/mapstructure"
)

func ReadVDF[T any](fn string) (v T, err error) {
	f, err := os.Open(fn)
	if err != nil {
		return v, fmt.Errorf("ReadVDF: %w", err)
	}
	defer f.Close()

	m, err := vdf.NewParser(f).Parse()
	if err != nil {
		return v, fmt.Errorf("ReadVDF: %w", err)
	}

	if err := mapstructure.Decode(m, &v); err != nil {
		return v, fmt.Errorf("ReadVDF: %w", err)
	}
	return v, nil
}
