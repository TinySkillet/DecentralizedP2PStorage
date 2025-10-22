package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

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
		case rpc := <-s.Transport.Consume():
			var msg Message
			err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg)
			if err != nil {
				log.Println(err)
			}

			fmt.Printf("Received: %s\n", string(msg.Payload.([]byte)))

			peer, found := s.peers[rpc.From]
			if !found {
				panic("Peer not found in peers map")
			}

			b := make([]byte, 1028)
			if _, err := peer.Read(b); err != nil {
				panic(err)
			}

			fmt.Printf("Received payload: %s\n", string(b))

			// TODO: make an separate interface instead of casting to TCPPeer
			peer.(*p2p.TCPPeer).Wg.Done()

		case <-s.quitch:
			return
		}
	}
}

// func (s *FileServer) handleMessage(msg *Message) error {
// 	// switch v := msg.Payload.(type) {

// 	// }
// 	// return nil
// }

func (s *FileServer) broadcast(msg *Message) error {

	// Peer implements net.Conn which implements Writer interface
	// therefore we can use Peer as a writer
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) StoreData(key string, r io.Reader) error {

	buf := new(bytes.Buffer)

	msg := Message{
		Payload: []byte("storagekey"),
	}

	err := gob.NewEncoder(buf).Encode(msg)
	if err != nil {
		return err
	}

	// send the message first
	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(time.Second * 2)

	// send the payload
	payload := []byte("THIS LARGE FILE")
	for _, peer := range s.peers {
		if err := peer.Send(payload); err != nil {
			return err
		}
	}

	return nil
	// // tee reads from r and writes everything it reads into buf
	// tee := io.TeeReader(r, buf)

	// // store file to disk
	// // as we read from tee, tee reads from r and writes same data to buf
	// if err := s.store.Write(key, tee); err != nil {
	// 	return err
	// }

	// _, err := io.Copy(buf, r)
	// if err != nil {
	// 	return err
	// }

	// p := &Payload{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }

	// // broadcast file to all known peers in the network
	// return s.broadcast(p)
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

type Message struct {
	Payload any
}
