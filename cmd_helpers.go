package main

import (
	"context"
	"path/filepath"
	"strings"

	dbpkg "github.com/TinySkillet/DecentralizedP2PStorage/db"
	"github.com/TinySkillet/DecentralizedP2PStorage/p2p"
)

// getStorageRoot returns the storage root directory based on listen address
// Uses a sanitized version of the port for the folder name
func getStorageRoot(listenAddr string) string {
	// Extract port from listen address (e.g., ":3000" -> "3000")
	port := strings.TrimPrefix(listenAddr, ":")
	if strings.Contains(port, ":") {
		// Handle "host:port" format
		parts := strings.Split(port, ":")
		port = parts[len(parts)-1]
	}
	return "node_" + port + "_data"
}

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
		StorageRoot:       getStorageRoot(listenAddr),
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer

	return s
}

func makeServerWithDB(listenAddr string, db *dbpkg.DB, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	// Use the database directory as base for storage root
	storageRoot := getStorageRoot(listenAddr)
	if db != nil {
		// If we have a DB, use its directory as the storage base
		dbDir := filepath.Dir(db.Path())
		if dbDir != "." && dbDir != "" {
			storageRoot = filepath.Join(dbDir, "files")
		}
	}

	fileServerOpts := FileServerOpts{
		EncryptionKey:     newEcryptionKey(),
		PathTransformFunc: CASPathTransformFunc,
		StorageRoot:       storageRoot,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
		DB:                db,
	}
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}

func loadOrInitKey(d *dbpkg.DB) ([]byte, error) {
	return d.GetOrCreateDefaultKey(context.Background(), newEcryptionKey)
}
