package data

import (
	"encoding/base64"
	"io"
	"unsafe"
)

func URI(mime string, s []byte) string {
	hdr := "data:" + mime + ";base64,"
	dat := make([]byte, len(hdr)+base64.StdEncoding.EncodedLen(len(s)))
	copy(dat, hdr)
	base64.StdEncoding.Encode(dat[len(hdr):], s)
	return unsafe.String(unsafe.SliceData(dat), len(dat))
}

func ReadURI(mime string, r io.Reader) (string, error) {
	s, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return URI(mime, s), nil
}
