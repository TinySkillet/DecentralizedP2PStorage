package p2p

import (
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type DefaultDecoder struct{}

func (decoder DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	buf := make([]byte, 2028)

	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	msg.Payload = buf[:n]

	return nil
}
