package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
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
	
	for _, peerInfo := range peers {
		if attempted >= maxAttempts {
			fmt.Printf("[%s] Reached max discovery attempts (%d), stopping\n", myAddr, maxAttempts)
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
			fmt.Printf("[%s] Already connected to %s, skipping\n", myAddr, peerInfo.Address)
			continue
		}
		
		// Attempt connection
		fmt.Printf("[%s] Attempting to connect to discovered peer %s\n", myAddr, peerInfo.Address)
		
		err := s.Transport.Dial(peerInfo.Address)
		if err != nil {
			fmt.Printf("[%s] Failed to connect to %s: %v\n", myAddr, peerInfo.Address, err)
		} else {
			fmt.Printf("[%s] Successfully connected to discovered peer %s\n", myAddr, peerInfo.Address)
			attempted++
			// Small delay to avoid overwhelming the network
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	fmt.Printf("[%s] Peer discovery complete. Connected to %d new peers\n", myAddr, attempted)
}

// sendPeerExchange sends our known peer list to a specific peer.
func (s *FileServer) sendPeerExchange(peerAddr string) error {
	fmt.Printf("[%s] DEBUG: sendPeerExchange called for peer %s\n", s.Transport.Address(), peerAddr)
	
	// Only send if we have database configured
	if s.DB == nil {
		fmt.Printf("[%s] DEBUG: Database is nil, skipping peer exchange\n", s.Transport.Address())
		return nil
	}
	
	fmt.Printf("[%s] DEBUG: Getting active peers from database...\n", s.Transport.Address())
	
	// Get active peers from database (seen in last 30 minutes, max 50 peers)
	activePeers, err := s.DB.GetActivePeers(context.Background(), 30*time.Minute, 50)
	if err != nil {
		fmt.Printf("[%s] Error getting active peers: %v\n", s.Transport.Address(), err)
		return err
	}
	
	fmt.Printf("[%s] DEBUG: Got %d active peers from database\n", s.Transport.Address(), len(activePeers))
	
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
	
	fmt.Printf("[%s] Sending peer exchange with %d peers to %s\n", s.Transport.Address(), len(peerInfos), peerAddr)
	
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
	return peer.Send(buf.Bytes())
}
