package p2p

// Peer represents the remote node.
type Peer interface {
	Close() error
}

// Transport handles communication between nodes.
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
