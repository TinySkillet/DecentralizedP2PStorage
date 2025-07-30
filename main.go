package main

import (
	"log"

	"github.com/TinySkillet/DecentralizedP2PStorage/p2p"
)

func main() {

	transport := p2p.NewTCPTransport(":3000")

	err := transport.ListenAndAccept()
	if err != nil {
		log.Fatal(err)
	}

	select {}
}
