package main

import (
	"log"

	"github.com/TinySkillet/DecentralizedP2PStorage/p2p"
)

func OnPeer(peer p2p.Peer) error {
	log.Printf("Doing some logic with the peer outside of tcp transport\n")
	peer.Close()
	return nil
}

func main() {

	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: ":3000",
		ShakeHands: p2p.NOPHandshakeFunc,
		Decoder:    p2p.DefaultDecoder{},
		OnPeer:     OnPeer,
	}
	transport := p2p.NewTCPTransport(tcpOpts)

	go func() {
		for {
			msg := <-transport.Consume()
			log.Println(string(msg.Payload))
		}
	}()

	err := transport.ListenAndAccept()
	if err != nil {
		log.Fatal(err)
	}

	select {}
}
