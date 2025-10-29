package main

import (
	"bytes"
	"fmt"
	"io"
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
		EncryptionKey:     newEcryptionKey(),
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
	s3 := makeServer(":5000", ":3000", ":4000")

	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(1 * time.Second)

	go func() {
		log.Fatal(s2.Start())
	}()
	time.Sleep(1 * time.Second)

	go s3.Start()
	time.Sleep(1 * time.Second)

	key := "coolpicture.jpg"
	data := bytes.NewReader([]byte("my big data file here!"))
	s3.Store(key, data)

	if err := s3.store.Delete(key); err != nil {
		log.Fatal(err)
	}

	_, r, err := s3.Get(key)
	if err != nil {
		log.Fatal(err)
	}

	// time.Sleep(3 * time.Millisecond)

	// _, r, err := s2.Get("coolpicture.jpg")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	b, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}
