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
		log.Println("File server stopped due to error or user quit action")
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg)
			if err != nil {
				log.Println("Decoding error: ", err)
			}

			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("Error while handling message: ", err)
			}

			// fmt.Printf("Received: %v\n", msg.Payload)

			// peer, found := s.peers[rpc.From]
			// if !found {
			// 	panic("Peer not found in peers map")
			// }

			// b := make([]byte, 1028)
			// if _, err := peer.Read(b); err != nil {
			// 	panic(err)
			// }

			// fmt.Printf("Received payload: %s\n", string(b))

			// // TODO: make an separate interface instead of casting to TCPPeer
			// peer.(*p2p.TCPPeer).Wg.Done()

		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, found := s.peers[from]
	if !found {
		return fmt.Errorf("peer (%s) could not be found in the peers list", from)
	}

	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	log.Printf("written %d bytes to disk\n", n)

	peer.(*p2p.TCPPeer).Wg.Done()
	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("need to serve file %s but it does not exist on disk", msg.Key)
	}

	fmt.Printf("Serving file %s over the network\n", msg.Key)
	r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not in map", from)
	}

	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("Written %d bytes over the network to %s\n", n, from)
	return nil
}

func (s *FileServer) stream(msg *Message) error {

	// Peer implements net.Conn which implements Writer interface
	// therefore we can use Peer as a writer
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		fmt.Printf("File %s found!\n", key)
		return s.store.Read(key)
	}

	fmt.Printf("Dont have file %s locally, searching on network...\n", key)

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	for _, peer := range s.peers {
		fmt.Println("Receiving stream from peer: ", peer.RemoteAddr().String())
		fileBuf := new(bytes.Buffer)
		n, err := io.Copy(fileBuf, peer)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Received %d bytes over the network\n", n)
		fmt.Println(fileBuf.String())
	}

	select {}

	// return nil, nil
}

func (s *FileServer) Store(key string, r io.Reader) error {

	fileBuf := new(bytes.Buffer)
	tee := io.TeeReader(r, fileBuf)

	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Second * 2)

	// TODO: use a multiwriter
	for _, peer := range s.peers {
		// then send the payload
		peer.Send([]byte{p2p.IncomingStream})
		n, err := io.Copy(peer, fileBuf)
		if err != nil {
			return err
		}

		fmt.Println("Received and written bytes to disk: ", n)
	}

	return nil

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

// register MessageStoreFile on gob, since we use any for payload
func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
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

type MessageStoreFile struct {
	Key  string
	Size int64
}

type MessageGetFile struct {
	Key string
}
