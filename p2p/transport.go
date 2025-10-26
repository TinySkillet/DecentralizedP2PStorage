package p2p

import "net"

// Peer represents the remote node.
type Peer interface {
	//interface embedding
	net.Conn
	Send([]byte) error
	CloseStream()
}

// Transport handles communication between nodes.
type Transport interface {
	Address() string
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
