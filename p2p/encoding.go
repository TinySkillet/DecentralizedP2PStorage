package p2p

import (
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type DefaultDecoder struct{}

func (decoder DefaultDecoder) Decode(r io.Reader, msg *RPC) error {

	peekBuf := make([]byte, 1)
	if _, err := r.Read(peekBuf); err != nil {
		return err
	}

	// in case of a stream, we do not decode
	stream := peekBuf[0] == IncomingStream
	if stream {
		msg.Stream = true
		return nil
	}

	buf := make([]byte, 2028)

	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	msg.Payload = buf[:n]

	return nil
}
