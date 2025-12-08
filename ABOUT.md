# Internal Architecture & Logic Documentation

This document provides a comprehensive deep-dive into the internal workings of the Decentralized P2P Storage system. It details the architectural decisions, protocol logic, data structures, and algorithms that power the application.

---

## 1. High-Level Architecture

The system is designed as a **hybrid unstructured P2P network** with **Content-Addressable Storage (CAS)**.

*   **Hybrid P2P**: While fully decentralized in operation (peers communicate directly), it supports "bootstrap nodes" to help new nodes join the network. Once joined, nodes form a mesh network.
*   **Content-Addressable Storage**: Files are not stored by filename but by the SHA-1 hash of their content (or key). This ensures data integrity and deduplication.
*   **Encrypted Storage**: All data is encrypted at rest using AES encryption. Data is transmitted in plaintext to allow peers to verify content and re-encrypt with their own keys.
*   **Local-First Database**: Each node maintains its own SQLite database for metadata, peer tracking, and file indexing.

---

## 2. Core Components

### 2.1. Transport Layer (`p2p/tcp_transport.go`)
The foundation of the network is a custom TCP transport layer.

*   **Connection Handling**:
    *   **Inbound**: The `startAcceptLoop` accepts incoming TCP connections.
    *   **Outbound**: The `Dial` method initiates connections to other peers.
*   **RPC Protocol**:
    *   Communication is message-based using a custom `RPC` struct.
    *   **Streaming**: The protocol supports a `Stream` flag. When true, the connection switches from message-passing mode to raw stream mode. This is critical for transferring large files without loading them entirely into memory.
    *   **Gob Encoding**: Go's `encoding/gob` is used for efficient binary serialization of messages.

### 2.2. Storage Engine (`storage.go`)
The storage engine manages files on the local disk.

*   **Path Transformation (CAS)**:
    *   Files are stored using a `PathTransformFunc`.
    *   **Logic**: The key (e.g., "my-file") is hashed using SHA-1. The hash is split into blocks (e.g., `abcde/f1234/...`) to create a deep directory structure.
    *   **Why?**: This prevents file system limits on the number of files in a single directory and ensures even distribution.
*   **Encryption**:
    *   **Encrypted at Rest**: Files are encrypted *before* writing to disk using `WriteEncrypt`.
    *   **Plaintext on Wire**: Files are decrypted *while* reading from disk using `ReadDecrypt` before being sent over the network. This allows each node to manage its own encryption keys independently.

### 2.3. Database Layer (`db/repo.go`, `db/db.go`)
Each node runs an embedded SQLite database to track state.

*   **Schema**:
    *   `peers`: Tracks known nodes (`address`, `last_seen`, `status`). Used for gossip and bootstrapping.
    *   `files`: Metadata for stored files (`hash`, `size`, `local_path`).
    *   `shares`: Tracks which peers have copies of which files (replication tracking).
    *   `keys`: Stores encryption keys (AES).
51: 
52: ### 2.4. Concurrency & Journaling
53: To ensure high performance and reliability under concurrent load (e.g., simultaneous uploads/downloads), the database uses **Write-Ahead Logging (WAL)**.
54: 
55: *   **WAL Mode**: Enabled via `PRAGMA journal_mode=WAL`. This allows readers to proceed without blocking writers, significantly improving concurrency.
56: *   **Shared Memory (.shm)**: When WAL is enabled, SQLite creates a `.shm` file. This memory-mapped file is used as an index for the WAL file, allowing multiple processes to access the log efficiently.
57: *   **Write-Ahead Log (.wal)**: This file contains transactions that have not yet been checkpointed to the main `.db` file.
58: *   **Busy Timeout**: A 5000ms `busy_timeout` is configured to gracefully handle transient locking contention.

---

## 3. Protocol Logic & Workflows

### 3.1. Peer Discovery (Gossip Protocol)
The system uses a proactive gossip protocol to maintain network connectivity.

**The Logic:**
1.  **Bootstrapping**: A node starts with a list of `BootstrapNodes`. It connects to them immediately.
2.  **Handshake & Exchange**: Upon connection (`OnPeer`), the node:
    *   Records the peer in the DB.
    *   Queries its own DB for "Active Peers" (seen in last 30 mins).
    *   Sends a `MessagePeerExchange` containing this list to the new peer.
3.  **Discovery**: When a node receives a `MessagePeerExchange`:
    *   It parses the list of peers.
    *   Filters out itself and already connected peers.
    *   Attempts to dial the new peers (rate-limited to prevent connection storms).
4.  **Result**: The network topology naturally expands. Connecting to one node eventually connects you to its neighbors.

### 3.2. File Storage (Upload Workflow)
Uploading a file triggers a broadcast and replication process.

**The Logic:**
1.  **Local Write**: The file is first written to the local node's storage (encrypted).
2.  **Metadata**: A record is inserted into the `files` table.
3.  **Broadcast**: A `MessageStoreFile` (containing key and size) is broadcast to **all connected peers**.
4.  **Streaming**:
    *   Peers receive the message and prepare to accept a stream.
    *   The uploader opens a stream to *all* peers simultaneously (`MultiWriter`).
    *   The file data is read from disk (encrypted) -> decrypted -> network (plaintext) -> peer encrypts -> peer disk (encrypted).
    *   **Framing**: Messages are length-prefixed to ensure correct stream handling over TCP.
5.  **Replication Tracking**: The uploader records `shares` in the DB, noting which peers received the file.

### 3.3. File Retrieval (Download Workflow)
Retrieving a file involves searching the network.

**The Logic:**
1.  **Local Check**: The node first checks if the file exists locally. If yes, it serves it immediately.
2.  **Network Search**: If not found, it broadcasts a `MessageGetFile` to all peers.
3.  **Peer Response**:
    *   Peers check their local storage.
    *   If a peer has the file, it opens a stream back to the requester.
    *   If a peer has the file, it opens a stream back to the requester.
    *   It sends the plaintext size, followed by the raw plaintext bytes (decrypted from its local storage).
4.  **Download & Decrypt**: The requester receives the stream, encrypts it on-the-fly (if storing) or writes it to the output file.
    *   For `get` command: Network (plaintext) -> Output File.
    *   For replication: Network (plaintext) -> Encrypt -> Disk.
5.  **First-Come-First-Served**: The first peer to respond effectively serves the file (logic currently handles sequential responses).

### 3.4. File Deletion
Deletion is a distributed operation ensuring consistency.

**The Logic:**
1.  **Local Delete**: The file is removed from the local disk and the `files` table.
2.  **Broadcast**: A `MessageDeleteFile` is sent to all connected peers.
3.  **Peer Action**:
    *   Peers receive the request.
    *   They attempt to delete the file from their `files` table.
    *   They attempt to delete the file from their disk (checking both original key and hash).
    *   **Consistency**: If the DB delete fails, the disk delete is skipped to prevent "ghost" records.

---

## 4. Internal Data Structures

### 4.1. Message Types (`server.go`)
*   `MessageStoreFile`: `{Key, Size}` - Announces incoming file transfer.
*   `MessageGetFile`: `{Key}` - Requests a file.
*   `MessageDeleteFile`: `{Key}` - Requests deletion.
*   `MessagePeerExchange`: `{Peers[]}` - Contains list of `PeerInfo`.

### 4.2. PeerInfo Struct
*   `Address`: IP:Port string.
*   `LastSeen`: Timestamp.

---

## 5. Security & Reliability

### 5.1. Encryption
*   **Algorithm**: AES (Advanced Encryption Standard).
*   **Implementation**: Stream cipher mode (likely CTR or OFB based on `copyEncrypt` usage) allows encrypting/decrypting streams of arbitrary size without loading the whole file into RAM.
*   **Key Management**: Keys are generated and stored in the `keys` table.

### 5.2. Reliability Mechanisms
*   **Atomic Database Operations**: File metadata and disk operations are synchronized.
*   **Stale Peer Cleanup**: The `cleanup` command removes peers inactive for >1 hour to keep the routing table healthy.
*   **Error Filtering**: The networking layer intelligently filters "broken pipe" errors caused by clients disconnecting after successful operations, keeping logs clean.

---

## 6. Comparison Points for Proposal

| Feature | Implementation Detail |
| :--- | :--- |
| **Decentralization** | No central server; every node is equal (peer). |
| **Scalability** | Gossip protocol ensures nodes find each other without a central registry. |
| **Efficiency** | Streaming I/O prevents memory exhaustion for large files. |
| **Privacy** | Encryption at rest ensures node operators cannot read stored data. (Note: Data is transmitted in plaintext). |
| **Persistence** | SQLite database ensures metadata survives restarts. |

---

## 7. System Integration & Deployment Architecture

The application is designed to integrate seamlessly with Linux system services, following modern deployment practices.

### 7.1. Service Daemonization
The application runs as a foreground process, delegating daemonization and process management to `systemd`.
*   **Parameterized Service**: The `p2p-storage@.service` uses systemd's template feature. The `@` symbol allows multiple independent instances to run on the same machine (e.g., `p2p-storage@user1`), isolating data and permissions per user.
*   **Security Hardening**: The service definition implements strict security controls:
    *   `NoNewPrivileges=true`: Prevents the process from escalating privileges.
    *   `ProtectSystem=strict`: Mounts the entire file system hierarchy read-only.
    *   `ReadWritePaths`: Explicitly grants write access *only* to the specific data directory (`~/.p2p`), minimizing the attack surface.

### 7.2. Configuration Strategy
The system implements a tiered configuration logic to support both development and production environments:
1.  **CLI Flags**: Highest priority. Used for overriding settings during manual testing or specific operations.
2.  **Config File**: Loaded from `~/.p2p/config`. This is the primary method for production configuration, allowing persistent settings for bootstrap nodes and ports.
3.  **Defaults**: Fallback values (e.g., port `:3000`) ensure the application works out-of-the-box for local testing.

