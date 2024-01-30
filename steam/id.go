package steam

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/amitybell/srcvox/errs"
)

const (
	// see also https://developer.valvesoftware.com/wiki/SteamID
	idMagic = 76561197960265728
)

var (
	_ json.Marshaler   = ID(0)
	_ json.Unmarshaler = (*ID)(nil)
)

type ID uint32

func (i ID) To64() uint64 {
	return uint64(i) + idMagic
}

func (i ID) To32() uint32 {
	return uint32(i)
}

func (i ID) mashal() []byte {
	return strconv.AppendUint(nil, uint64(i.To32()), 10)
}

func (i ID) String32() string {
	return strconv.FormatUint(uint64(i.To32()), 10)
}

func (i ID) String64() string {
	return strconv.FormatUint(i.To64(), 10)
}

func (i ID) String() string {
	return string(i.mashal())
}

func (i ID) MarshalJSON() ([]byte, error) {
	return i.mashal(), nil
}

func (i *ID) UnmarshalJSON(p []byte) error {
	var err error
	*i, err = ParseID(string(p))
	return err
}

func ToID(n uint64) ID {
	if n <= idMagic {
		return ID(n)
	}
	return ID(n - idMagic)
}

func ParseID(s string) (id ID, err error) {
	defer errs.Recover(&err)

	if n, err := strconv.ParseUint(s, 10, 64); err == nil {
		return ToID(n), nil
	}

	a := strings.Split(strings.Trim(s, `[]`), `:`)
	if len(a) != 3 {
		return 0, fmt.Errorf("Invalid steam id format: %s", s)
	}
	y, err := strconv.ParseUint(a[1], 10, 64)
	if err != nil {
		return 0, err
	}
	z, err := strconv.ParseUint(a[2], 10, 64)
	if err != nil {
		return 0, err
	}
	return ToID(z*2 + y), err
}
