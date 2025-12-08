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
			if rpc.Stream {
				if err := s.handleStream(rpc.From); err != nil {
					log.Printf("[%s] Error handling stream: %v", s.Transport.Address(), err)
				}
				continue
			}

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
	case MessagePeerExchange:
		return s.handleMessagePeerExchange(from, v)
	}
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	fmt.Printf("[%s] Received StoreFile message from %s for key %s. Expecting stream...\n", s.Transport.Address(), from, msg.Key)
	s.pendingFileTransfers[from] = msg
	return nil
}

func (s *FileServer) handleStream(from string) error {
	// Check for pending upload (StoreFile)
	if msg, ok := s.pendingFileTransfers[from]; ok {
		delete(s.pendingFileTransfers, from)

		peer, found := s.peers[from]
		if !found {
			return fmt.Errorf("peer (%s) could not be found in the peers list", from)
		}

		// Receive plaintext, write encrypted
		n, err := s.store.WriteEncrypt(s.EncryptionKey, msg.Key, io.LimitReader(peer, msg.Size))
		if err != nil {
			return err
		}

		fmt.Printf("[%s] Written %d bytes to disk (encrypted) from %s\n", s.Transport.Address(), n, from)

		// Record share in database if configured
		if s.DB != nil {
			shareID := hashKey(msg.Key + from + "incoming")
			_ = s.DB.InsertShare(context.Background(), dbpkg.Share{
				ID:        shareID,
				FileID:    msg.Key,
				PeerID:    from,
				Direction: "incoming",
			})
		}

		peer.CloseStream()

		// Signal download completion if anyone is waiting
		if ch, ok := s.downloadChannels[msg.Key]; ok {
			close(ch)
			delete(s.downloadChannels, msg.Key)
		}

		return nil
	}

	// Check for pending download (GetFile)
	// We need to know WHICH file is being downloaded.
	// But handleStream only knows 'from'.
	// We need a map[from]chan struct{} for downloads too?
	// Or map[from]expectedKey?

	// Since Get is synchronous, we can assume only one download per peer at a time?
	// Or we can use pendingDownloads map[string]chan struct{} where key is... peer?
	// But Get broadcasts to ALL peers.
	// Any peer might respond.

	// If we use pendingDownloads map[string]chan struct{} where key is fileKey (hash).
	// But handleStream doesn't know fileKey!

	// Wait, if we use MessageStoreFile for downloads too (as planned),
	// then the responder sends MessageStoreFile FIRST.
	// So it ends up in pendingFileTransfers!
	// So the logic above handles it automatically!

	// The only difference is that for Get, we might want to signal the Get caller.
	// But Get caller can just check if file exists?
	// Or we can have a separate notification mechanism.

	// If Get caller waits for file to exist, we are good.

	return fmt.Errorf("peer %s sent a stream but no pending transfer was found", from)
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	fmt.Printf("[%s] Received request to serve file '%s'\n", s.Transport.Address(), msg.Key)

	// Try to find the file.
	// It could be stored with original key (if we are uploader) or hashed key (if we are peer).

	keyToRead := msg.Key

	if s.DB != nil {
		// Check if we have the original key mapping
		// We need a way to look up file by hash.
		// ListFiles is inefficient but works for now.
		// Ideally we should have GetFileByHash.
		files, err := s.DB.ListFiles(context.Background())
		if err == nil {
			for _, f := range files {
				if f.Hash == msg.Key {
					fmt.Printf("[%s] Found original key '%s' for hash '%s'\n", s.Transport.Address(), f.Name, msg.Key)
					keyToRead = f.Name
					break
				}
			}
		}
	}

	// Read from disk (Encrypted) -> Decrypt -> Send Plaintext
	// We use ReadDecrypt to get plaintext reader

	// Use s.store.Read which is public
	encSize, r, err := s.store.Read(keyToRead)
	if err != nil {
		// If failed with original key (or default hashed key), try the other one?
		// If we switched to original key, maybe we should fallback?
		// But if DB says we have it, we should have it.
		return fmt.Errorf("[%s] Failed to read file %s: %v", s.Transport.Address(), keyToRead, err)
	}
	if rc, ok := r.(io.ReadCloser); ok {
		rc.Close()
	}

	// Note: ReadDecrypt returns 0 size because it's a stream.
	// We need to know the size to send MessageStoreFile.
	// We can get encrypted size from DB or disk, and subtract 16.

	plaintextSize := encSize - 16

	// Now open decrypt stream
	_, fileReader, err := s.store.ReadDecrypt(s.EncryptionKey, keyToRead)
	if err != nil {
		return err
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not found in peer list", from)
	}

	// Send MessageStoreFile first (acting as uploader)
	storeMsg := Message{
		Payload: MessageStoreFile{
			Key:  msg.Key,
			Size: plaintextSize,
		},
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(&storeMsg); err != nil {
		return err
	}

	peer.Send([]byte{p2p.IncomingMessage})
	binary.Write(peer, binary.LittleEndian, int64(buf.Len()))
	if err := peer.Send(buf.Bytes()); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	// send the 'IncomingStream' byte to the peer first
	peer.Send([]byte{p2p.IncomingStream})

	n, err := io.Copy(peer, fileReader)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] Written %d bytes (plaintext) over the network to %s\n", s.Transport.Address(), n, from)
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

	// Delete from database if configured
	// Note: For peer messages, we log warnings but continue with file deletion
	// to honor the delete request, even if DB cleanup fails
	dbDeleteFailed := false
	if s.DB != nil {
		if err := s.DB.DeleteFile(context.Background(), msg.Key); err != nil {
			fmt.Printf("[%s] WARNING: Failed to delete file with hash '%s' from database: %v. Continuing with file deletion - DATABASE INCONSISTENCY DETECTED\n", s.Transport.Address(), msg.Key, err)
			dbDeleteFailed = true
		} else {
			fmt.Printf("[%s] Deleted file with hash '%s' from database\n", s.Transport.Address(), msg.Key)
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

	if dbDeleteFailed {
		fmt.Printf("[%s] WARNING: Database inconsistency - file deleted from disk but database cleanup failed for hash '%s'\n", s.Transport.Address(), msg.Key)
	}
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
		binary.Write(peer, binary.LittleEndian, int64(buf.Len()))
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
		return s.store.ReadDecrypt(s.EncryptionKey, key)
	}

	fmt.Printf("[%s] Did not find file '%s' locally, searching on network...\n", s.Transport.Address(), key)

	// Create channel to wait for download
	hash := hashKey(key)
	ch := make(chan struct{})
	s.downloadChannels[hash] = ch

	msg := Message{
		Payload: MessageGetFile{
			Key: hash,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		delete(s.downloadChannels, hash)
		return 0, nil, err
	}

	// Wait for download to complete or timeout
	select {
	case <-ch:
		fmt.Printf("[%s] File downloaded successfully!\n", s.Transport.Address())
		// The file was downloaded and stored using the hash
		return s.store.ReadDecrypt(s.EncryptionKey, hash)
	case <-time.After(10 * time.Second):
		delete(s.downloadChannels, hash)
		return 0, nil, fmt.Errorf("timeout waiting for file download")
	}
}

func (s *FileServer) Store(key string, r io.Reader) error {

	// 1. Write Encrypted to disk.
	n, err := s.store.WriteEncrypt(s.EncryptionKey, key, r)
	if err != nil {
		return err
	}

	plaintextSize := n - 16

	// Record file metadata if DB is configured
	if s.DB != nil {
		_ = s.DB.InsertFileWithKey(context.Background(), dbpkg.File{
			ID:        hashKey(key),
			Name:      key,
			Hash:      hashKey(key),
			Size:      plaintextSize,
			LocalPath: s.store.FullPathForKey(key),
		}, "default")
	}

	// Capture peers
	s.peersLock.Lock()
	peers := []io.Writer{}
	peerAddrs := []string{}
	for addr, peer := range s.peers {
		peers = append(peers, peer)
		peerAddrs = append(peerAddrs, addr)
	}
	s.peersLock.Unlock()

	msg := Message{
		Payload: MessageStoreFile{
			Key:  hashKey(key),
			Size: plaintextSize,
		},
	}

	// Broadcast to captured peers
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(&msg); err != nil {
		return err
	}

	for i, peer := range peers {
		addr := peerAddrs[i]
		fmt.Printf("[%s] Sending message to peer %s\n", s.Transport.Address(), addr)
		if p, ok := peer.(p2p.Peer); ok {
			p.Send([]byte{p2p.IncomingMessage})
			binary.Write(p, binary.LittleEndian, int64(buf.Len()))
			if err := p.Send(buf.Bytes()); err != nil {
				fmt.Printf("[%s] Error sending message to peer %s: %v\n", s.Transport.Address(), addr, err)
			}
		}
	}

	// 3. Stream Plaintext to peers.
	// Read from disk (Encrypted) -> Decrypt -> Send.
	_, fileReader, err := s.store.ReadDecrypt(s.EncryptionKey, key)
	if err != nil {
		return err
	}

	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream})

	// Copy plaintext to peers
	written, err := io.Copy(mw, fileReader)
	if err != nil {
		return err
	}

	// Record shares for each peer that received the file
	if s.DB != nil {
		fileID := hashKey(key)
		for _, addr := range peerAddrs {
			shareID := hashKey(fileID + addr + "outgoing")
			_ = s.DB.InsertShare(context.Background(), dbpkg.Share{
				ID:        shareID,
				FileID:    fileID,
				PeerID:    addr,
				Direction: "outgoing",
			})
		}
	}

	fmt.Printf("[%s] Received and written %d bytes to disk (encrypted), sent %d bytes (plaintext) to peers\n", s.Transport.Address(), n, written)

	return nil
}

func (s *FileServer) Delete(key string) error {
	// Delete from database if configured
	if s.DB != nil {
		fileID := hashKey(key)
		if err := s.DB.DeleteFile(context.Background(), fileID); err != nil {
			// Fail fast to maintain consistency: if DB delete fails, don't delete the file
			return fmt.Errorf("[%s] Failed to delete file '%s' from database: %v. File not deleted from disk to maintain consistency", s.Transport.Address(), key, err)
		}
		fmt.Printf("[%s] Deleted file '%s' from database\n", s.Transport.Address(), key)
	}

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
	peerAddr := p.RemoteAddr().String()
	fmt.Printf("[%s] Connected with remote %s\n", s.Transport.Address(), peerAddr)

	if s.DB != nil {
		now := time.Now()
		_ = s.DB.UpsertPeer(context.Background(), dbpkg.Peer{
			ID:       peerAddr,
			Address:  peerAddr,
			Status:   "connected",
			LastSeen: &now,
		})
	}

	// Send peer exchange to newly connected peer (after lock is released)
	go func() {
		// Small delay to ensure connection is fully established
		time.Sleep(500 * time.Millisecond)

		// Retry a few times if database is locked
		for i := 0; i < 5; i++ {
			if err := s.sendPeerExchange(peerAddr); err != nil {
				// Only log unexpected errors (filtering done in sendPeerExchange)
				fmt.Printf("[%s] Error sending peer exchange to %s: %v (attempt %d/5)\n", s.Transport.Address(), peerAddr, err, i+1)
				time.Sleep(1 * time.Second)
				continue
			}
			break
		}
	}()

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
	gob.Register(MessagePeerExchange{})
	gob.Register(PeerInfo{})
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

	pendingFileTransfers map[string]MessageStoreFile
	downloadChannels     map[string]chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts:       opts,
		store:                NewStore(storeOpts),
		quitch:               make(chan struct{}),
		peers:                make(map[string]p2p.Peer),
		pendingFileTransfers: make(map[string]MessageStoreFile),
		downloadChannels:     make(map[string]chan struct{}),
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

type MessagePeerExchange struct {
	Peers []PeerInfo
}

type PeerInfo struct {
	Address  string
	LastSeen time.Time
}
