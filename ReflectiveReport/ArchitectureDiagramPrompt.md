# Architecture Diagram Prompt for Eraser.io

## Decentralized P2P Storage System Architecture

Create a balanced architectural diagram of a peer-to-peer file storage system:

### Node Structure (Show one node, then network view)

**Core Components per Node:**
1. **CLI** - User interface (store, get, delete commands)
2. **FileServer** - Orchestrates operations, manages peers, handles messages
3. **Transport** - TCP networking (connections, message framing, streams)
4. **Storage (CAS)** - Content-addressable storage with SHA-256 hashing
5. **Database** - SQLite metadata (key→hash mapping, peer info)

**Key Features:**
- Files stored by SHA-256 content hash (not filename)
- AES-256 encryption before network transfer
- Full replication: all peers store all files

### Network View
- 2 nodes connected via TCP: Node A (:3000) and Node B (:4000)
- Bidirectional TCP connection between them
- Show message exchange between nodes

### Operation Flows with Messages

**Store:**
- Node A: CLI → FileServer (compute SHA-256 hash) → Storage (write by hash) → Database (store key→hash mapping)
- Node A: FileServer encrypts → Broadcasts MessageStoreFile{hash, size} → Streams encrypted data
- Node B: Receives MessageStoreFile → Receives encrypted stream → Decrypts → Storage (write by hash)

**Get:**
- Node A: CLI → FileServer → Database (lookup key→hash) → Check Storage
- If missing: Node A broadcasts MessageGetFile{hash}
- Node B: Receives MessageGetFile → Checks Storage → Streams file data
- Node A: Receives data → Decrypts → Storage (write) → Returns to CLI

**Delete:**
- Node A: CLI → FileServer → Database (lookup key→hash) → Delete from Storage & Database
- Node A: Broadcasts MessageDeleteFile{hash}
- Node B: Receives MessageDeleteFile → Deletes from Storage & Database

### Message Types
- MessageStoreFile: {Key: contentHash, Size: int64}
- MessageGetFile: {Key: contentHash}
- MessageDeleteFile: {Key: contentHash}
- Messages encoded with GOB (Go binary encoding)


