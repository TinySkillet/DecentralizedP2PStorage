package main

import (
	"log"
	"time"

	"github.com/TinySkillet/DecentralizedP2PStorage/p2p"
)

func OnPeer(peer p2p.Peer) error {
	log.Printf("Doing some logic with the peer outside of tcp transport\n")
	peer.Close()
	return nil
}

func main() {

	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tcpTransport := p2p.NewTCPTransport(tcpOpts)

	fileServerOpts := FileServerOpts{
		PathTransformFunc: CASPathTransformFunc,
		StorageRoot:       "3000_network",
		Transport:         tcpTransport,
	}

	s := NewFileServer(fileServerOpts)

	go func() {
		time.Sleep(time.Second * 3)
		s.Stop()
	}()

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}

}
