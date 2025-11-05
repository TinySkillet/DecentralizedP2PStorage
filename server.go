package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"context"

	dbpkg "github.com/TinySkillet/DecentralizedP2PStorage/db"
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
		log.Printf("[%s] File server stopped due to error or user quit action\n", s.Transport.Address())
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg)
			if err != nil {
				log.Printf("[%s] Decoding error: %v", s.Transport.Address(), err)
			}

			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Printf("[%s] Error while handling message: %v\n", s.Transport.Address(), err)
			}

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
	case MessageDeleteFile:
		return s.handleMessageDeleteFile(from, v)
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

	fmt.Printf("[%s] Written %d bytes to disk\n", s.Transport.Address(), n)

	peer.CloseStream()
	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	fmt.Printf("[%s] Received request to serve file '%s'\n", s.Transport.Address(), msg.Key)

	if !s.store.Has(msg.Key) {
		return fmt.Errorf("[%s] Received request to serve file %s but it does not exist on disk", s.Transport.Address(), msg.Key)
	}

	fmt.Printf("[%s] Serving file '%s' over the network\n", s.Transport.Address(), msg.Key)

	size, r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		defer rc.Close()
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not found in peer list", from)
	}

	// send the 'IncomingStream' byte to the peer first
	peer.Send([]byte{p2p.IncomingStream})

	// then we can send the file size
	binary.Write(peer, binary.LittleEndian, size)

	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] Written %d bytes over the network to %s\n", s.Transport.Address(), n, from)
	return nil
}

func (s *FileServer) handleMessageDeleteFile(from string, msg MessageDeleteFile) error {
	fmt.Printf("[%s] Received delete request for file with hash '%s' from %s\n", s.Transport.Address(), msg.Key, from)

	// The msg.Key is the hashed key. Files can be stored in two ways:
	// 1. Locally stored with original key (metadata in DB, file stored with hashed path)
	// 2. Received from peer with hashed key (no metadata, file stored with double-hashed path)
	// We need to try both approaches.

	var originalKey string
	if s.DB != nil {
		// Try to find the file by hash in the database to get the original key
		files, err := s.DB.ListFiles(context.Background())
		if err == nil {
			for _, f := range files {
				if f.Hash == msg.Key {
					originalKey = f.Name
					break
				}
			}
		}
	}

	// First, try to delete using original key (if file was stored locally)
	if originalKey != "" {
		if s.store.Has(originalKey) {
			if err := s.store.Delete(originalKey); err != nil {
				return fmt.Errorf("[%s] Error deleting file '%s': %v", s.Transport.Address(), originalKey, err)
			}
			fmt.Printf("[%s] Deleted file '%s' from local storage\n", s.Transport.Address(), originalKey)
			return nil
		}
	}

	// If original key approach didn't work, try deleting using the hashed key directly
	// (for files received from peers, which were stored with the hashed key)
	if s.store.Has(msg.Key) {
		if err := s.store.Delete(msg.Key); err != nil {
			return fmt.Errorf("[%s] Error deleting file with hash '%s': %v", s.Transport.Address(), msg.Key, err)
		}
		fmt.Printf("[%s] Deleted file with hash '%s' from local storage\n", s.Transport.Address(), msg.Key)
		return nil
	}

	fmt.Printf("[%s] File with hash '%s' does not exist locally, skipping deletion\n", s.Transport.Address(), msg.Key)
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

	s.peersLock.Lock()
	defer s.peersLock.Unlock()

	for addr, peer := range s.peers {
		fmt.Printf("[%s] Sending message to peer %s\n", s.Transport.Address(), addr)
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil {
			fmt.Printf("[%s] Error sending message to peer %s: %v\n", s.Transport.Address(), addr, err)
			return err
		}
	}

	return nil
}

func (s *FileServer) Get(key string) (int64, io.Reader, error) {
	if s.store.Has(key) {
		fmt.Printf("[%s] File '%s' found locally! Serving file from disk...\n", s.Transport.Address(), key)
		return s.store.Read(key)
	}

	fmt.Printf("[%s] Did not find file '%s' locally, searching on network...\n", s.Transport.Address(), key)

	msg := Message{
		Payload: MessageGetFile{
			Key: hashKey(key),
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return 0, nil, err
	}

	time.Sleep(time.Millisecond * 500)

	for _, peer := range s.peers {
		// first read the file size so we can limit the amt of bytes
		// that we read from the connection, so it will not keep hanging

		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)

		n, err := s.store.WriteDecrypt(s.EncryptionKey, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return 0, nil, err
		}

		fmt.Printf("[%s] Received %d bytes over the network from [%s]\n", s.Transport.Address(), n, peer.RemoteAddr())
		peer.CloseStream()
	}

	return s.store.Read(key)

}

func (s *FileServer) Store(key string, r io.Reader) error {

	fileBuf := new(bytes.Buffer)
	tee := io.TeeReader(r, fileBuf)

	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	// Record file metadata if DB is configured
	if s.DB != nil {
		_ = s.DB.InsertFileWithKey(context.Background(), dbpkg.File{
			ID:        hashKey(key),
			Name:      key,
			Hash:      hashKey(key),
			Size:      size,
			LocalPath: s.store.FullPathForKey(key),
		}, "default")
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  hashKey(key),
			Size: size + 16, // IV which is 16 bytes is prepended
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)

	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream})
	n, err := copyEncrypt(s.EncryptionKey, fileBuf, mw)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] Received and written %d bytes to disk\n", s.Transport.Address(), n)

	return nil
}

func (s *FileServer) Delete(key string) error {
	// Delete locally first
	if !s.store.Has(key) {
		fmt.Printf("[%s] File '%s' does not exist locally\n", s.Transport.Address(), key)
		// Still broadcast the delete in case other peers have it
	} else {
		if err := s.store.Delete(key); err != nil {
			return err
		}
		fmt.Printf("[%s] Deleted file '%s' from local storage\n", s.Transport.Address(), key)
	}

	// Check if we have any peers connected
	s.peersLock.Lock()
	peerCount := len(s.peers)
	peerAddrs := make([]string, 0, len(s.peers))
	for addr := range s.peers {
		peerAddrs = append(peerAddrs, addr)
	}
	s.peersLock.Unlock()

	fmt.Printf("[%s] Connected to %d peer(s): %v\n", s.Transport.Address(), peerCount, peerAddrs)

	if peerCount == 0 {
		fmt.Printf("[%s] No peers connected, cannot broadcast delete message\n", s.Transport.Address())
		return nil
	}

	// Broadcast delete message to all peers
	msg := Message{
		Payload: MessageDeleteFile{
			Key: hashKey(key),
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	fmt.Printf("[%s] Broadcasted delete request for '%s' to %d peer(s)\n", s.Transport.Address(), key, peerCount)
	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

// in OnPeer
func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peersLock.Lock()
	defer s.peersLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	fmt.Printf("[%s] Connected with remote %s\n", s.Transport.Address(), p.RemoteAddr().String())

	if s.DB != nil {
		now := time.Now()
		_ = s.DB.UpsertPeer(context.Background(), dbpkg.Peer{
			ID:       p.RemoteAddr().String(),
			Address:  p.RemoteAddr().String(),
			Status:   "connected",
			LastSeen: &now,
		})
	}
	return nil
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			fmt.Printf("[%s] Attempting to connect with remote: %s\n", s.Transport.Address(), addr)

			err := s.Transport.Dial(addr)
			if err != nil {
				fmt.Printf("[%s] Dial error: %v\n", s.Transport.Address(), err)
			}
		}(addr)
	}
	return nil
}

// waitForPeers waits for at least one peer connection, with a timeout
func (s *FileServer) waitForPeers(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		s.peersLock.Lock()
		peerCount := len(s.peers)
		s.peersLock.Unlock()

		if peerCount > 0 {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for peer connections")
}

// register MessageStoreFile on gob, since we use any for payload
func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
	gob.Register(MessageDeleteFile{})
}

type FileServerOpts struct {
	EncryptionKey     []byte
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
	DB                *dbpkg.DB
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

type MessageDeleteFile struct {
	Key string
}
