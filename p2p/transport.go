package p2p

import "net"

// Peer represents the remote node.
type Peer interface {
	//interface embedding
	net.Conn
	Close() error
	Send([]byte) error
}

// Transport handles communication between nodes.
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
