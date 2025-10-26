package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// Implements the Transport interface
func (t TCPTransport) Close() error {
	return t.listener.Close()
}

// Implements the Transport interface
func (t *TCPTransport) Address() string {
	return t.ListenAddr
}

// Consume implements the Transport interface, which returns a read only channel for
// reading incoming messages from another peer
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChan
}

// Implements the Transport interface
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)
	return nil
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
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("TCP accept error: %s\n", err)
		}

		log.Printf("New Incoming Connection: %+v\n", conn.RemoteAddr().String())
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		log.Printf("Dropping peer connection: %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read Loop
	for {
		rpc := RPC{}
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}

		rpc.From = conn.RemoteAddr().String()

		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("Incoming stream from [%s], waiting till stream is done...\n", conn.RemoteAddr().String())

			peer.wg.Wait()
			fmt.Printf("Stream from [%s] closed. Resuming normal read loop.\n", conn.RemoteAddr().String())

			continue
		}

		t.rpcChan <- rpc
	}
}

// TCPPeer represents the remote node over a TCP connection.
type TCPPeer struct {
	// underlying connection of the peer
	// in this case is a tcp connection
	net.Conn

	// If we dial and retrieve a conn: outbound = true.
	// If we accept and retrieve a conn:  outbound = false.
	outbound bool

	wg *sync.WaitGroup
}

// implements the Peer interface
func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

// implements the Peer interface
func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcChan  chan RPC
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcChan:          make(chan RPC, 1024),
	}
}
