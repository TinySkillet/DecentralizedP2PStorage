package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strings"
	"time"

	"github.com/TinySkillet/DecentralizedP2PStorage/p2p"
)

// handleMessagePeerExchange handles incoming peer lists from other nodes.
// This enables peer discovery beyond bootstrap nodes.
func (s *FileServer) handleMessagePeerExchange(from string, msg MessagePeerExchange) error {
	fmt.Printf("[%s] Received peer exchange with %d peers from %s\n", s.Transport.Address(), len(msg.Peers), from)
	
	// Discover and connect to new peers
	go s.discoverPeers(msg.Peers)
	
	return nil
}

// discoverPeers attempts to connect to newly discovered peers.
// It filters out duplicates and our own address, and limits connection attempts.
func (s *FileServer) discoverPeers(peers []PeerInfo) {
	myAddr := s.Transport.Address()
	maxAttempts := 10 // Limit to prevent connection storms
	attempted := 0
	connected := 0
	
	for _, peerInfo := range peers {
		if attempted >= maxAttempts {
			break
		}
		
		// Skip if it's our own address
		if peerInfo.Address == myAddr {
			continue
		}
		
		// Skip if already connected
		s.peersLock.Lock()
		_, alreadyConnected := s.peers[peerInfo.Address]
		s.peersLock.Unlock()
		
		if alreadyConnected {
			continue
		}
		
		// Attempt connection
		err := s.Transport.Dial(peerInfo.Address)
		if err == nil {
			fmt.Printf("[%s] Connected to discovered peer %s\n", myAddr, peerInfo.Address)
			connected++
			attempted++
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	if connected > 0 {
		fmt.Printf("[%s] Peer discovery: connected to %d new peer(s)\n", myAddr, connected)
	}
}

// sendPeerExchange sends our known peer list to a specific peer.
func (s *FileServer) sendPeerExchange(peerAddr string) error {
	// Only send if we have database configured
	if s.DB == nil {
		return nil
	}
	
	// Get active peers from database (seen in last 30 minutes, max 50 peers)
	activePeers, err := s.DB.GetActivePeers(context.Background(), 30*time.Minute, 50)
	if err != nil {
		fmt.Printf("[%s] Error getting active peers: %v\n", s.Transport.Address(), err)
		return err
	}
	
	// Convert to PeerInfo slice
	peerInfos := make([]PeerInfo, 0, len(activePeers))
	for _, p := range activePeers {
		if p.LastSeen != nil {
			peerInfos = append(peerInfos, PeerInfo{
				Address:  p.Address,
				LastSeen: *p.LastSeen,
			})
		}
	}
	
	fmt.Printf("[%s] Sending %d peer(s) to %s\n", s.Transport.Address(), len(peerInfos), peerAddr)
	
	// Send to specific peer
	msg := Message{
		Payload: MessagePeerExchange{
			Peers: peerInfos,
		},
	}
	
	// Find the peer connection
	s.peersLock.Lock()
	peer, ok := s.peers[peerAddr]
	s.peersLock.Unlock()
	
	if !ok {
		return fmt.Errorf("peer %s not found in connected peers", peerAddr)
	}
	
	// Encode and send
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	
	peer.Send([]byte{p2p.IncomingMessage})
	err = peer.Send(buf.Bytes())
	
	// Filter out expected errors from short-lived connections
	if err != nil && !isExpectedNetworkError(err) {
		// Only log unexpected errors
		return err
	}
	
	return nil
}

// isExpectedNetworkError checks if an error is an expected network condition
// that doesn't need to be reported (e.g., client disconnected after completing their request)
func isExpectedNetworkError(err error) bool {
	if err == nil {
		return false
	}
	
	errMsg := err.Error()
	
	// Expected errors from short-lived connections (store/get/delete commands)
	expectedErrors := []string{
		"broken pipe",                    // Client already disconnected
		"use of closed network connection", // Connection already closed
		"connection reset by peer",       // Client forcefully closed
		"EOF",                           // Clean disconnect
	}
	
	for _, expected := range expectedErrors {
		if strings.Contains(errMsg, expected) {
			return true
		}
	}
	
	return false
}
