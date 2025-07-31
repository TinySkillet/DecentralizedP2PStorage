package main

import (
	"log"

	"github.com/TinySkillet/DecentralizedP2PStorage/p2p"
)

func main() {

	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: ":3000",
		ShakeHands: p2p.NOPHandshakeFunc,
	}
	transport := p2p.NewTCPTransport(tcpOpts)

	err := transport.ListenAndAccept()
	if err != nil {
		log.Fatal(err)
	}

	select {}
}
