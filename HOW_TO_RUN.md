# How to Run - P2P Decentralized Storage

## Prerequisites

```bash
# Build the binary
cd /home/tinyskillet/Src/P2PStorage
go build
```

---

## Commands Reference

### 1. `serve` - Start a P2P Node

**Purpose:** Start a persistent P2P storage node that listens for connections and participates in the network.

**Syntax:**
```bash
./DecentralizedP2PStorage serve [flags]
```

**Flags:**
| Flag | Default | Description |
|------|---------|-------------|
| `--listen` | `:3000` | Address to listen on (format: `:PORT` or `HOST:PORT`) |
| `--db` | `p2p.db` | SQLite database path |
| `--bootstrap` | (none) | Comma-separated bootstrap node addresses |
| `--config` | (none) | Config file path (overrides defaults, CLI flags override config) |

**Examples:**

```bash
# Start a bootstrap node (first node in network)
mkdir -p node_3000
./DecentralizedP2PStorage serve --listen :3000 --db node_3000/p2p.db

# Start a node that connects to existing network
mkdir -p node_4000
./DecentralizedP2PStorage serve --listen :4000 --db node_4000/p2p.db --bootstrap localhost:3000

# Start with config file
./DecentralizedP2PStorage serve --config ~/.p2p/config
```

**Expected Behavior:**
- Creates database if it doesn't exist
- Connects to bootstrap nodes if specified
- Exchanges peer lists with connected peers (peer gossip)
- Automatically discovers and connects to other peers
- Listens for incoming connections indefinitely
- Stores files received from other peers

**Output:**
```
[:3000] Attempting to connect with remote: localhost:4000
Listening on TCP at PORT :3000
[:3000] Connected with remote [::1]:4000
[:3000] Sending 2 peer(s) to [::1]:4000
[:3000] Connected to discovered peer [::1]:5000
[:3000] Peer discovery: connected to 1 new peer(s)
```

**Important:** Each node needs its own database file. Using the same database for multiple nodes will cause "database is locked" errors.

---

### 2. `store` - Upload a File

**Purpose:** Store a file in the P2P network. Broadcasts the file to all connected peers.

**Syntax:**
```bash
./DecentralizedP2PStorage store <key> <file> [flags]
```

**Arguments:**
| Argument | Description |
|----------|-------------|
| `key` | Unique identifier for the file (used for retrieval) |
| `file` | Path to the file to store |

**Flags:**
| Flag | Default | Description |
|------|---------|-------------|
| `--listen` | `:3000` | Address to listen on (must be unique, not used by running nodes) |
| `--db` | `p2p.db` | SQLite database path |
| `--bootstrap` | (none) | Node(s) to connect to for broadcasting |

**Examples:**

```bash
# Store a file via a running node
echo "Hello World" > test.txt
./DecentralizedP2PStorage store myfile test.txt --listen :7000 --db /tmp/store.db --bootstrap localhost:3000

# Store with multiple bootstrap nodes
./DecentralizedP2PStorage store backup document.pdf --listen :7000 --db /tmp/store.db --bootstrap localhost:3000,localhost:4000
```

**Expected Behavior:**
1. Connects to specified bootstrap node(s)
2. Stores file locally (encrypted)
3. Broadcasts file to all connected peers
4. Exits after broadcast is complete

**Output (on running nodes):**
```
[:3000] Incoming stream from [sender], waiting till stream is done...
[:3000] Written 54 bytes to disk
[:3000] Stream from [sender] closed. Resuming normal read loop.
```

**Important:** 
- Use a unique port (e.g., `:7000`) that isn't used by running nodes
- Use a temporary database (`/tmp/store.db`) to avoid conflicts
- File is stored with the key you specify, not the filename

---

### 3. `get` - Download a File

**Purpose:** Retrieve a file from the P2P network using its key.

**Syntax:**
```bash
./DecentralizedP2PStorage get <key> [flags]
```

**Arguments:**
| Argument | Description |
|----------|-------------|
| `key` | The key used when storing the file |

**Flags:**
| Flag | Default | Description |
|------|---------|-------------|
| `--listen` | `:3000` | Address to listen on (must be unique) |
| `--db` | `p2p.db` | SQLite database path |
| `--bootstrap` | (none) | Node(s) to request the file from |
| `--out` | (stdout) | Output file path |

**Examples:**

```bash
# Retrieve a file
./DecentralizedP2PStorage get myfile --listen :8000 --db /tmp/get.db --bootstrap localhost:3000 --out retrieved.txt

# Verify content
cat retrieved.txt
```

**Expected Behavior:**
1. Connects to specified bootstrap node(s)
2. Requests file by key from connected peers
3. Receives file if available on any peer
4. Writes to output file or stdout
5. Exits after retrieval

**Output (on serving nodes):**
```
[:3000] Received request to serve file 'abc123...'
```

**Error Cases:**
- If file doesn't exist: `Error: file does not exist on disk`
- If key not found: No response from peers

---

### 4. `delete` - Remove a File

**Purpose:** Delete a file from the local node and broadcast deletion to all peers.

**Syntax:**
```bash
./DecentralizedP2PStorage delete <key> [flags]
```

**Arguments:**
| Argument | Description |
|----------|-------------|
| `key` | The key of the file to delete |

**Flags:**
| Flag | Default | Description |
|------|---------|-------------|
| `--listen` | `:3000` | Address to listen on |
| `--db` | `p2p.db` | SQLite database path |
| `--bootstrap` | (none) | Node(s) to broadcast deletion to |

**Examples:**

```bash
./DecentralizedP2PStorage delete myfile --listen :9000 --db /tmp/delete.db --bootstrap localhost:3000
```

**Expected Behavior:**
1. Connects to specified bootstrap node(s)
2. Broadcasts delete request to all peers
3. Each peer removes the file from disk and database
4. Exits after broadcast

**Output (on receiving nodes):**
```
[:3000] Received delete request for file with hash 'abc123...' from [sender]
[:3000] Deleted file with hash 'abc123...' from database
[:3000] Deleted file with hash 'abc123...' from disk
```

---

### 5. `peers` - List Known Peers

**Purpose:** Display all peers known to a node (from the database).

**Syntax:**
```bash
./DecentralizedP2PStorage peers [flags]
```

**Flags:**
| Flag | Default | Description |
|------|---------|-------------|
| `--db` | `p2p.db` | SQLite database path |

**Examples:**

```bash
# List peers for a specific node
./DecentralizedP2PStorage peers --db node_3000/p2p.db
```

**Expected Output:**
```
ADDRESS                       STATUS         LAST SEEN
----------------------------------------------------------------------
[::1]:4000                    connected      2025-12-07 22:20:45
[::1]:5000                    connected      2025-12-07 22:19:30
[::1]:43080                   connected      2025-12-07 22:18:15
```

**Notes:**
- Shows peers seen in the last 24 hours
- Status reflects last known state (may be stale)
- Ephemeral ports (e.g., `:43080`) are from short-lived client connections

---

### 6. `cleanup` - Remove Stale Peers

**Purpose:** Delete peer records that haven't been seen in the last hour.

**Syntax:**
```bash
./DecentralizedP2PStorage cleanup [flags]
```

**Flags:**
| Flag | Default | Description |
|------|---------|-------------|
| `--db` | `p2p.db` | SQLite database path |

**Examples:**

```bash
./DecentralizedP2PStorage cleanup --db node_3000/p2p.db
```

**Expected Output:**
```
Removed 7 stale peer(s)
```

**Use Case:** Run periodically to clean up ephemeral port entries and disconnected peers.

---

### 7. `files list` - Show Stored Files

**Purpose:** List all files stored in this node's database.

**Syntax:**
```bash
./DecentralizedP2PStorage files list [flags]
```

**Flags:**
| Flag | Default | Description |
|------|---------|-------------|
| `--db` | `p2p.db` | SQLite database path |

**Examples:**

```bash
./DecentralizedP2PStorage files list --db node_3000/p2p.db
```

---

### 8. `shares` - View File Shares

**Purpose:** Show which files have been shared with which peers.

**Syntax:**
```bash
./DecentralizedP2PStorage shares [flags]
```

**Flags:**
| Flag | Default | Description |
|------|---------|-------------|
| `--db` | `p2p.db` | SQLite database path |

**Examples:**

```bash
./DecentralizedP2PStorage shares --db node_3000/p2p.db
```

---

## Testing Scenarios

### Scenario 1: Peer Gossip Discovery

**Goal:** Verify that Node C automatically discovers Node A through Node B.

**Setup:**
```bash
# Clean up
rm -rf node_3000 node_4000 node_5000
mkdir -p node_3000 node_4000 node_5000

# Terminal 1: Start Node A (bootstrap)
./DecentralizedP2PStorage serve --listen :3000 --db node_3000/p2p.db

# Terminal 2: Start Node B (connects to A)
./DecentralizedP2PStorage serve --listen :4000 --db node_4000/p2p.db --bootstrap localhost:3000

# Terminal 3: Start Node C (connects to B only)
./DecentralizedP2PStorage serve --listen :5000 --db node_5000/p2p.db --bootstrap localhost:4000
```

**Expected Result (Terminal 3):**
```
[:5000] Connected with remote [::1]:4000
[:5000] Sending 1 peer(s) to [::1]:4000
[:5000] Connected to discovered peer [::1]:3000
[:5000] Peer discovery: connected to 1 new peer(s)
```

**Verification:**
```bash
./DecentralizedP2PStorage peers --db node_5000/p2p.db
# Should show both :3000 and :4000
```

---

### Scenario 2: File Replication

**Goal:** Store a file on one node and retrieve it from another.

**Prerequisite:** 3 nodes running (from Scenario 1)

**Steps:**
```bash
# Create test file
echo "P2P Storage Works!" > test.txt

# Store via Node C
./DecentralizedP2PStorage store testfile test.txt --listen :7000 --db /tmp/store.db --bootstrap localhost:5000

# Wait a moment for replication, then retrieve from Node A
./DecentralizedP2PStorage get testfile --listen :8000 --db /tmp/get.db --bootstrap localhost:3000 --out retrieved.txt

# Verify
cat retrieved.txt
# Output: P2P Storage Works!
```

---

### Scenario 3: Delete Propagation

**Goal:** Delete a file and verify it's removed from all nodes.

**Prerequisite:** File stored (from Scenario 2)

**Steps:**
```bash
# Delete file via Node A
./DecentralizedP2PStorage delete testfile --listen :9000 --db /tmp/delete.db --bootstrap localhost:3000

# Try to retrieve (should fail)
./DecentralizedP2PStorage get testfile --listen :9001 --db /tmp/verify.db --bootstrap localhost:5000 --out should_fail.txt
```

---

## File System Structure

When running a node with `--db node_3000/p2p.db`:

```
node_3000/
├── p2p.db          # SQLite database (peers, files, keys, shares)
└── files/          # Actual file storage (content-addressed)
    └── abc12/
        └── 3def.../
            └── abc123def... (file data, encrypted)
```

---

## Config File Format

Location: `~/.p2p/config` or specified via `--config`

```ini
# Listen address
listen=:3000

# Database path
db=~/.p2p/p2p.db

# Bootstrap nodes (comma-separated)
bootstrap=192.168.1.100:3000,192.168.1.101:3000
```

**Precedence:** CLI flags > Config file > Defaults

---

## Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `address already in use` | Port conflict | Use different `--listen` port |
| `database is locked` | Multiple processes using same DB | Use separate `--db` paths |
| `unable to open database file` | Parent directory doesn't exist | Create directory with `mkdir -p` |
| `connection refused` | Target node not running | Start the target node first |
| `file does not exist on disk` | File not stored on that node | Try different bootstrap node |

---

## Systemd Service (Production)

### Installation

```bash
# Run the install script
./install.sh
```

**What it does:**
1. Builds the binary
2. Copies binary to `/usr/local/bin/DecentralizedP2PStorage`
3. Creates config directory `~/.p2p/`
4. Creates default config file `~/.p2p/config`
5. Installs systemd service file `/etc/systemd/system/p2p-storage@.service`

### Configure Bootstrap Nodes

Edit the config file before starting the service:

```bash
nano ~/.p2p/config
```

**Config file (`~/.p2p/config`):**
```ini
# Listen address
listen=:3000

# Database path
db=~/.p2p/p2p.db

# Bootstrap nodes (leave empty for bootstrap node, or specify peers)
bootstrap=192.168.1.100:3000,192.168.1.101:3000
```

### Service Management

```bash
# Start the service
sudo systemctl start p2p-storage@$USER

# Stop the service
sudo systemctl stop p2p-storage@$USER

# Restart the service
sudo systemctl restart p2p-storage@$USER

# Check status
sudo systemctl status p2p-storage@$USER

# Enable on boot
sudo systemctl enable p2p-storage@$USER

# Disable on boot
sudo systemctl disable p2p-storage@$USER

# View logs (live)
sudo journalctl -u p2p-storage@$USER -f

# View last 100 log lines
sudo journalctl -u p2p-storage@$USER -n 100
```

### Service File Location

```
/etc/systemd/system/p2p-storage@.service
```

**Contents:**
```ini
[Unit]
Description=P2P Decentralized Storage Node
After=network.target

[Service]
Type=simple
User=%i
WorkingDirectory=/home/%i
ExecStart=/usr/local/bin/DecentralizedP2PStorage serve --config /home/%i/.p2p/config
Restart=on-failure
RestartSec=10

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/home/%i/.p2p

[Install]
WantedBy=multi-user.target
```

### Customizing the Service

To change listen port or other settings:

1. **Edit the config file** (recommended):
   ```bash
   nano ~/.p2p/config
   # Change listen=:3000 to listen=:4000
   sudo systemctl restart p2p-storage@$USER
   ```

2. **Or edit the service file** (for advanced changes):
   ```bash
   sudo nano /etc/systemd/system/p2p-storage@.service
   sudo systemctl daemon-reload
   sudo systemctl restart p2p-storage@$USER
   ```

### Uninstallation

```bash
# Run the uninstall script
./uninstall.sh
```

**Or manually:**
```bash
# Stop and disable service
sudo systemctl stop p2p-storage@$USER
sudo systemctl disable p2p-storage@$USER

# Remove service file
sudo rm /etc/systemd/system/p2p-storage@.service
sudo systemctl daemon-reload

# Remove binary
sudo rm /usr/local/bin/DecentralizedP2PStorage

# Remove data (optional)
rm -rf ~/.p2p
```

### Multiple Nodes on Same Machine

To run multiple nodes as services on the same machine, create separate config files:

```bash
# Create configs for each node
cp ~/.p2p/config ~/.p2p/config-node1
cp ~/.p2p/config ~/.p2p/config-node2

# Edit each config with different ports
# config-node1: listen=:3000, db=~/.p2p/node1.db
# config-node2: listen=:4000, db=~/.p2p/node2.db, bootstrap=localhost:3000

# Create separate service files
sudo cp /etc/systemd/system/p2p-storage@.service /etc/systemd/system/p2p-node1@.service
sudo cp /etc/systemd/system/p2p-storage@.service /etc/systemd/system/p2p-node2@.service

# Edit each to use different config files
# ExecStart=... --config /home/%i/.p2p/config-node1
# ExecStart=... --config /home/%i/.p2p/config-node2

sudo systemctl daemon-reload
sudo systemctl start p2p-node1@$USER
sudo systemctl start p2p-node2@$USER
```

---

## Cleanup

```bash
# Remove test data
rm -rf node_3000 node_4000 node_5000
rm -rf /tmp/store.db /tmp/get.db /tmp/delete.db /tmp/verify.db
rm -f test.txt retrieved.txt should_fail.txt
```
