package p2p

import (
	"log"
	"net"
	"sync"
)

// TCPPeer represents the remote node over a TCP connection.
type TCPPeer struct {
	// underlying connection of the peer
	conn net.Conn

	// If we dial and retrieve a conn: outbound = true.
	// If we accept and retrieve a conn:  outbound = false.
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{conn: conn, outbound: outbound}
}

type TCPTransport struct {
	listenAddress string
	listener      net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAddr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}
	t.listener = ln

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			log.Printf("TCP accept error: %s\n", err)
		}
		log.Printf("Listening on TCP at PORT %s\n", t.listenAddress)

		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	log.Printf("New Incoming Connection%+v\n", peer)
}
