package main

import (
	"bytes"
	"log"
	"time"

	"github.com/TinySkillet/DecentralizedP2PStorage/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		PathTransformFunc: CASPathTransformFunc,
		StorageRoot:       listenAddr + "_network",
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {

	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")

	go func() {
		err := s1.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(2 * time.Second)
	go s2.Start()
	time.Sleep(2 * time.Second)

	data := bytes.NewReader([]byte("my big data file here!"))
	s2.StoreData("myprivatedata", data)

	select {}
}
