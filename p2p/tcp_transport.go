package p2p

import (
	"errors"
	"io"
	"log"
	"net"
)

// TCPPeer represents the remote node over a TCP connection.
type TCPPeer struct {
	// underlying connection of the peer
	conn net.Conn

	// If we dial and retrieve a conn: outbound = true.
	// If we accept and retrieve a conn:  outbound = false.
	outbound bool
}

// Close implements the Peer interface.
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{conn: conn, outbound: outbound}
}

type TCPTransportOpts struct {
	ListenAddr string
	ShakeHands HandshakeFunc
	Decoder    Decoder
	OnPeer     func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcChan  chan RPC
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcChan:          make(chan RPC),
	}
}

// Consume implements the Transport interface, which returns a read only channel for
// reading incoming messages from another peer
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChan
}

func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	t.listener = ln

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	log.Printf("Listening on TCP at PORT %s\n", t.ListenAddr)
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			log.Printf("TCP accept error: %s\n", err)
		}

		log.Printf("New Incoming Connection%+v\n", conn.RemoteAddr().String())
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error

	defer func() {
		log.Printf("Dropping peer connection: %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, true)

	if err = t.ShakeHands(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read Loop
	rpc := RPC{}
	for {
		if err = t.Decoder.Decode(conn, &rpc); err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("%s disconnected!", conn.RemoteAddr().String())
				return
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("TCP error: %s\n", err)
			continue
		}
		rpc.From = conn.RemoteAddr()
		t.rpcChan <- rpc
	}
}
