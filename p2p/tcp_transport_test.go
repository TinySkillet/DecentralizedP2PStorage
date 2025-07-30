package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	tr := NewTCPTransport(":4000")

	assert.Equal(t, tr.listenAddress, ":4000")

	assert.Nil(t, tr.ListenAndAccept())
}
