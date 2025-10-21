package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/TinySkillet/DecentralizedP2PStorage/p2p"
)

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	if len(s.BootstrapNodes) != 0 {
		s.bootstrapNetwork()
	}

	s.loop()

	return nil
}

func (s *FileServer) loop() {

	defer func() {
		log.Println("File server stopped due to user quit action")
	}()

	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peersLock.Lock()
	defer s.peersLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("Connected with remote %s\n", p.RemoteAddr().String())
	return nil
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			fmt.Println("Attempting to connect with remote: ", addr)

			err := s.Transport.Dial(addr)
			if err != nil {
				log.Println("Dial error: ", err)
			}
		}(addr)
	}
	return nil
}

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peersLock sync.Mutex
	peers     map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}
