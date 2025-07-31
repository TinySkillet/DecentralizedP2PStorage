package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, any) error
}

type GOBDecoder struct{}

func (decoder GOBDecoder) Decode(r io.Reader, payload any) error {
	return gob.NewDecoder(r).Decode(payload)
}
