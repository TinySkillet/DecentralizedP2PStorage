# Decentralized P2P Storage

A decentralized peer-to-peer file storage system with automatic peer discovery and gossip protocol.

## Features

✨ **Automatic Peer Discovery** - Nodes discover each other through peer gossip protocol  
📦 **Distributed Storage** - Files replicated across multiple peers  
🔄 **Real-time Sync** - Automatic file synchronization  
🗄️ **SQLite Integration** - Persistent storage and peer tracking  
🚀 **Production Ready** - Systemd service, clean logging, easy deployment

---

## Quick Start

### Build

```bash
go build -o bin/p2p
```

### Local Development Setup

Run this script to create the necessary directories for the examples below:

```bash
./setup_local_nodes.sh
```

### Run a Node

```bash
# Node 1 (Bootstrap)
./bin/p2p serve --listen :3000 --db node_3000/p2p.db

# Node 2 (in another terminal)
./bin/p2p serve --listen :4000 --db node_4000/p2p.db --bootstrap localhost:3000

# Node 3 (in another terminal)
./bin/p2p serve --listen :5000 --db node_5000/p2p.db --bootstrap localhost:4000
```

Node 3 will automatically discover Node 1 through peer gossip! 🎉

---

## Commands

### Serve - Start a Node

````bash
```bash
./bin/p2p serve [flags]

Flags:
  --listen string        Listen address (default ":3000")
  --db string           Database path (default "p2p.db")
  --bootstrap strings   Bootstrap peer addresses
````

### Store - Upload a File

```bash
./bin/p2p store <key> <file> [flags]

# Example:
./bin/p2p store myfile document.pdf --listen :7000 --db /tmp/store.db --bootstrap localhost:3000
```

### Get - Download a File

```bash
./bin/p2p get <key> [flags]

# Example:
./bin/p2p get myfile --listen :7000 --db /tmp/get.db --bootstrap localhost:3000 --out retrieved.pdf
```

### Peers - List Known Peers

```bash
./bin/p2p peers --db <database-path>

# Example:
mkdir -p node_3000
./bin/p2p peers --db node_3000/p2p.db
```

**Output:**

```
ADDRESS                       STATUS         LAST SEEN
----------------------------------------------------------------------
[::1]:4000                    connected      2025-12-07 21:20:45
[::1]:5000                    connected      2025-12-07 21:19:30
```

### Cleanup - Remove Stale Peers

```bash
./bin/p2p cleanup --db <database-path>

# Example:
./bin/p2p cleanup --db node_3000/p2p.db
```

**Output:**

```
Removed 7 stale peer(s)
```

### Files List - Show Stored Files

```bash
./bin/p2p files list --db <database-path>

# Example:
./bin/p2p files list --db node_3000/p2p.db
```

### Delete - Remove a File

```bash
./bin/p2p delete <key> [flags]

# Example:
./bin/p2p delete myfile --listen :7000 --db /tmp/delete.db --bootstrap localhost:3000
```

### Shares - View File Replicas

```bash
./bin/p2p shares --db <database-path>

# Example:
./bin/p2p shares --db node_3000/p2p.db
```

---

## Testing Peer Gossip Protocol

### Automated Test

```bash
./test_peer_gossip.sh
```

This runs a 3-node test demonstrating automatic peer discovery.

### Manual Test

**Step 1: Start 3 nodes in separate terminals**

Terminal 1:

```bash
mkdir -p node_3000
mkdir -p node_3000
./bin/p2p serve --listen :3000 --db node_3000/p2p.db
```

Terminal 2:

```bash
mkdir -p node_4000
./bin/p2p serve --listen :4000 --db node_4000/p2p.db --bootstrap localhost:3000
```

Terminal 3:

```bash
mkdir -p node_5000
./bin/p2p serve --listen :5000 --db node_5000/p2p.db --bootstrap localhost:4000
```

**Step 2: Verify peer discovery**

Watch Terminal 3 - you should see:

```
[:5000] Connected to discovered peer [::1]:3000
[:5000] Peer discovery: connected to 1 new peer(s)
```

**Step 3: Check peer database**

```bash
./bin/p2p peers --db node_5000/p2p.db
```

You should see both `:3000` and `:4000`!

---

## Testing File Storage & Retrieval

### Store a File

```bash
# Create test file
echo "Hello P2P Storage!" > test.txt

# Store on Node C (port 5000)
# Store on Node C (port 5000)
# Note: We use a separate port (:7000) and DB to act as a client connecting to Node C
./bin/p2p store testfile test.txt --listen :7000 --db /tmp/store.db --bootstrap localhost:5000
```

### Retrieve from Another Node

```bash
# Retrieve from Node A (port 3000) - proving replication works!
./bin/p2p get testfile --listen :8000 --db /tmp/get.db --bootstrap localhost:3000 --out retrieved.txt

# Verify
cat retrieved.txt
# Output: Hello P2P Storage!
```

---

## Production Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for complete deployment guide including:

- Systemd service setup
- Firewall configuration
- Monitoring
- Troubleshooting

### Quick Install

```bash
./install.sh
```

### Run as Systemd Service

```bash
# Start
sudo systemctl start p2p-storage@$USER

# Enable on boot
sudo systemctl enable p2p-storage@$USER

# Check status
sudo systemctl status p2p-storage@$USER
```

---

## Testing Checklist

- [ ] **Peer Discovery**: Node C discovers Node A through Node B
- [ ] **File Storage**: Store file on one node
- [ ] **File Retrieval**: Retrieve file from different node
- [ ] **Peer List**: `peers` command shows connected nodes
- [ ] **Cleanup**: `cleanup` removes stale peers
- [ ] **Deletion**: Delete propagates to all nodes

### Run All Tests

```bash
# 1. Test peer discovery
./test_peer_gossip.sh

# 2. Create test directories
mkdir -p node_3000 node_4000 node_5000

# 3. Check peers (with running nodes)
./bin/p2p peers --db node_5000/p2p.db

# 4. Clean up stale peers
./bin/p2p cleanup --db node_5000/p2p.db

# 5. List files
./bin/p2p files list --db node_5000/p2p.db
```

---

## Architecture

```
┌─────────────┐         Gossip          ┌─────────────┐
│   Node A    │ ◄────────────────────── │   Node B    │
│   :3000     │                          │   :4000     │
└─────────────┘                          └─────────────┘
       ▲                                        │
       │                                        │
       │         Peer Discovery                 ▼
       │         (Automatic!)            ┌─────────────┐
       └─────────────────────────────────│   Node C    │
                                         │   :5000     │
                                         └─────────────┘
```

- **Peer Gossip**: Nodes exchange peer lists automatically
- **File Replication**: Files broadcast to all connected peers
- **SQLite Storage**: Persistent peer and file metadata
- **TCP Transport**: Reliable peer-to-peer connections

---

## Troubleshooting

### "unable to open database file: out of memory"

The database directory doesn't exist. Create it first:

```bash
mkdir -p node_3000
mkdir -p node_3000
./bin/p2p peers --db node_3000/p2p.db
```

### "address already in use"

Another process is using that port. Use a different port:

```bash
./bin/p2p serve --listen :3001 --db node_3001/p2p.db
```

### "database is locked"

Multiple instances trying to use same database. Use separate databases:

```bash
./bin/p2p serve --listen :3000 --db node_3000/p2p.db
./bin/p2p serve --listen :4000 --db node_4000/p2p.db
```

### Connection Refused Errors

These are normal for stale peer addresses. The system:

- Tries old addresses
- Skips failed connections
- Connects to active peers

Run cleanup to remove stale peers:

```bash
./bin/p2p cleanup --db node_3000/p2p.db
```

---

## Development

### Project Structure

```
.
├── cmd.go              # CLI commands
├── server.go           # File server logic
├── peer_exchange.go    # Peer gossip protocol
├── storage.go          # File storage layer
├── db/
│   ├── repo.go         # Database operations
│   └── schema.sql      # Database schema
├── p2p/
│   ├── tcp_transport.go    # TCP transport layer
│   ├── message.go          # Message encoding
│   └── encoding.go         # Data encoding
├── DEPLOYMENT.md       # Deployment guide
└── test_peer_gossip.sh # Automated test
```

### Build & Test

```bash
# Build
go build -o bin/p2p

# Run tests
go test ./...

# Test peer gossip
./test_peer_gossip.sh
```

---

## License

MIT License - See LICENSE file for details

---

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

---

## Acknowledgments

Built with:

- Go 1.x
- SQLite
- Cobra CLI framework
