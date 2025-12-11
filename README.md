# Decentralized P2P Storage

A decentralized peer-to-peer file storage system with automatic peer discovery and gossip protocol.

## Features

*   **Automatic Peer Discovery**: Nodes discover each other via gossip protocol.
*   **Distributed Storage**: Files are replicated across peers.
*   **Real-time Sync**: Automatic synchronization of files.
*   **SQLite Integration**: Persistent storage for metadata.
*   **Production Ready**: Includes systemd service integration.

## Build

```bash
go build -o bin/p2p
```

## Local Testing

Follow this flow to test the network locally with 3 nodes.

### 1. Setup

```bash
./setup_local_nodes.sh
```

### 2. Start Node 3000 (Bootstrap Node)

```bash
./bin/p2p serve --listen :3000 --db node_3000/p2p.db
```

### 3. Start Node 4000 (Connects to Node 3000)

Open a new terminal:

```bash
./bin/p2p serve --listen :4000 --db node_4000/p2p.db --bootstrap localhost:3000
```

### 4. Start Node 5000 (Connects to Node 4000)

Open a new terminal. We will use this node to perform operations.

```bash
./bin/p2p serve --listen :5000 --db node_5000/p2p.db --bootstrap localhost:4000
```

### 5. Operations (Using Node 5000)

**Store a file:**

```bash
# Create a dummy file
echo "Hello P2P World" > hello.txt

# Store it (using a temporary client port)
./bin/p2p store hello hello.txt --listen :6000 --db node_5000/p2p.db --bootstrap localhost:4000
```

**List files:**

```bash
./bin/p2p files list --db node_5000/p2p.db
```

**Get a file:**

```bash
./bin/p2p get hello --listen :6000 --db node_5000/p2p.db --bootstrap localhost:4000 --out retrieved_hello.txt
cat retrieved_hello.txt
```

**Delete a file:**

```bash
./bin/p2p delete hello --listen :6000 --db node_5000/p2p.db --bootstrap localhost:4000
```

**Check Peers:**

```bash
./bin/p2p peers --db node_5000/p2p.db
```

## Systemd Integration

### Installation

We provide an installation script to set up the P2P Storage node as a systemd service.

```bash
./install.sh
```

This will:
1.  Build the binary.
2.  Install it to `/usr/local/bin/DecentralizedP2PStorage`.
3.  Create a config directory at `~/.p2p`.
4.  Install the systemd service file.

### Service Management

**Start the service:**
```bash
sudo systemctl start p2p-storage@$USER
```

**Enable on boot:**
```bash
sudo systemctl enable p2p-storage@$USER
```

**Check status:**
```bash
sudo systemctl status p2p-storage@$USER
```

**View logs:**
```bash
sudo journalctl -u p2p-storage@$USER -f
```

### Service Settings

The service is configured via the `~/.p2p/config` file. You can edit this file to change the listen address, bootstrap nodes, and other settings.

The systemd service file is located at `/etc/systemd/system/p2p-storage@.service`. It runs the service as the specified user.

### Uninstallation

To remove the installation:

1.  **Stop and disable the service:**
    ```bash
    sudo systemctl stop p2p-storage@$USER
    sudo systemctl disable p2p-storage@$USER
    ```

2.  **Remove the service file:**
    ```bash
    sudo rm /etc/systemd/system/p2p-storage@.service
    sudo systemctl daemon-reload
    ```

3.  **Remove the binary:**
    ```bash
    sudo rm /usr/local/bin/DecentralizedP2PStorage
    ```

4.  **Remove configuration (Optional):**
    ```bash
    rm -rf ~/.p2p
    ```

