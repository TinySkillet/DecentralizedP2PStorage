# Decentralized P2P Storage

A decentralized peer-to-peer (P2P) file storage system written in Go. This system allows nodes (peers) to connect, share, and retrieve files from each other over a TCP network. Files are encrypted and stored using content-addressable storage (CAS) with automatic peer discovery and file replication.

## Features

- **Decentralized Storage**: Files are distributed across multiple peers in the network
- **Content-Addressable Storage (CAS)**: Files are stored based on their content hash
- **Encryption**: Files are encrypted using AES encryption
- **Peer Discovery**: Automatic connection to bootstrap nodes
- **File Operations**: Store, retrieve, and delete files across the network
- **SQLite Database**: Metadata tracking for files and peers
- **Command-Line Interface**: Easy-to-use CLI with Cobra

## Prerequisites

- **Go 1.24.5 or higher**: [Download Go](https://golang.org/dl/)
- **Git**: For cloning the repository (optional)

## Installation

### 1. Clone the Repository

```bash
git clone <repository-url>
cd P2PStorage
```

### 2. Install Dependencies

The project uses Go modules. Dependencies will be automatically downloaded when you build:

```bash
go mod download
```

### 3. Build the Project

You can build the project using either method:

**Using Make:**
```bash
make build
```

This will create a binary at `bin/p2p`.

**Using Go directly:**
```bash
go build -o bin/p2p
```

## Usage

The CLI tool provides several commands for managing your P2P storage node. All commands support a persistent `--db` flag to specify the SQLite database path (defaults to `p2p.db`).

### Global Flags

- `--db <path>`: Specify the SQLite database path (default: `p2p.db`)

### Commands

#### 1. Serve (Run a Node)

Start a P2P storage node that listens for connections and can serve files to other peers.

```bash
./bin/p2p serve [flags]
```

**Flags:**
- `--listen <address>`: Listen address (default: `:3000`)
- `--bootstrap <nodes>`: Bootstrap nodes to connect to (comma-separated or repeated flag)

**Examples:**

```bash
# Start a node on default port 3000
./bin/p2p serve

# Start a node on a custom port
./bin/p2p serve --listen :4000

# Start a node and connect to bootstrap nodes
./bin/p2p serve --listen :4000 --bootstrap :3000 --bootstrap :5000

# Use a custom database
./bin/p2p serve --db mynode.db
```

#### 2. Store (Store a File)

Store a file locally and broadcast it to all connected peers.

```bash
./bin/p2p store <key> <file> [flags]
```

**Arguments:**
- `key`: The key/name to store the file under
- `file`: Path to the file to store

**Flags:**
- `--listen <address>`: Listen address (default: `:3000`)
- `--bootstrap <nodes>`: Bootstrap nodes to connect to

**Examples:**

```bash
# Store a file
./bin/p2p store myfile.txt /path/to/file.txt

# Store a file and connect to bootstrap nodes
./bin/p2p store document.pdf ./doc.pdf --bootstrap :3000

# Store with custom listen address
./bin/p2p store image.jpg ./photo.jpg --listen :4000 --bootstrap :3000
```

#### 3. Get (Retrieve a File)

Fetch a file from the network (local storage or peers).

```bash
./bin/p2p get <key> [flags]
```

**Arguments:**
- `key`: The key/name of the file to retrieve

**Flags:**
- `--listen <address>`: Listen address (default: `:3000`)
- `--bootstrap <nodes>`: Bootstrap nodes to connect to
- `--out <path>`: Output file path (if not specified, outputs to stdout)

**Examples:**

```bash
# Get a file and output to stdout
./bin/p2p get myfile.txt

# Get a file and save to a specific location
./bin/p2p get myfile.txt --out ./downloaded.txt

# Get a file from the network
./bin/p2p get document.pdf --bootstrap :3000 --out ./doc.pdf
```

#### 4. Delete (Delete a File)

Delete a file from local storage.

```bash
./bin/p2p delete <key> [flags]
```

**Arguments:**
- `key`: The key/name of the file to delete

**Flags:**
- `--listen <address>`: Listen address (default: `:3000`)
- `--bootstrap <nodes>`: Bootstrap nodes to connect to

**Examples:**

```bash
# Delete a file
./bin/p2p delete myfile.txt

# Delete a file with custom database
./bin/p2p delete document.pdf --db mynode.db
```

#### 5. Files List

List all known files in the database.

```bash
./bin/p2p files list [flags]
```

**Output Format:**
```
ID    Name    Size    LocalPath
```

**Examples:**

```bash
# List all files
./bin/p2p files list

# List files with custom database
./bin/p2p files list --db mynode.db
```

#### 6. Demo (Run Local Demo)

Run a local 3-node demo to test the P2P storage system.

```bash
./bin/p2p demo
```

This command:
1. Starts three nodes on ports `:3000`, `:4000`, and `:5000`
2. Connects them to form a network
3. Stores a test file
4. Deletes the file
5. Retrieves the file from the network
6. Displays the file content

**Example:**

```bash
./bin/p2p demo
```

## Common Workflows

### Setting Up a Multi-Node Network

**Terminal 1 - Start Bootstrap Node:**
```bash
./bin/p2p serve --listen :3000
```

**Terminal 2 - Start Second Node:**
```bash
./bin/p2p serve --listen :4000 --bootstrap :3000
```

**Terminal 3 - Start Third Node:**
```bash
./bin/p2p serve --listen :5000 --bootstrap :3000 --bootstrap :4000
```

### Storing and Retrieving Files

**On Node 1 (port 3000):**
```bash
./bin/p2p serve --listen :3000
```

**On Node 2 (port 4000):**
```bash
# Start the node
./bin/p2p serve --listen :4000 --bootstrap :3000

# In another terminal, store a file
./bin/p2p store myfile.txt ./example.txt --listen :4000 --bootstrap :3000
```

**On Node 3 (port 5000):**
```bash
# Start the node
./bin/p2p serve --listen :5000 --bootstrap :3000

# Retrieve the file stored by Node 2
./bin/p2p get myfile.txt --listen :5000 --bootstrap :3000 --out ./retrieved.txt
```

## Project Structure

```
P2PStorage/
├── main.go              # Entry point
├── cmd.go               # CLI commands definition
├── cmd_helpers.go       # Helper functions for commands
├── server.go            # FileServer implementation
├── storage.go           # Storage layer with CAS
├── crypto.go            # Encryption utilities
├── db/
│   ├── db.go           # Database connection
│   └── repo.go         # Database operations
└── p2p/
    ├── transport.go     # Transport interface
    ├── tcp_transport.go # TCP transport implementation
    ├── message.go       # Message definitions
    ├── encoding.go     # Message encoding/decoding
    └── handshake.go    # Connection handshake
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
make test
```

### Building

```bash
# Build the binary
make build

# Or use Go directly
go build -o bin/p2p
```

## Database

The system uses SQLite to store:
- File metadata (ID, name, size, local path)
- Peer information (address, status, last seen)
- Encryption keys

By default, the database is stored as `p2p.db` in the current directory. You can specify a custom path using the `--db` flag.

## Storage

Files are stored using Content-Addressable Storage (CAS) in a directory structure based on the file key hash. The default storage root is `<listen_address>_network` (e.g., `:3000_network`).

Files are encrypted using AES encryption before storage.

## Troubleshooting

### Port Already in Use

If you get an error about a port being in use, choose a different port:

```bash
./bin/p2p serve --listen :4000
```

### Cannot Connect to Bootstrap Nodes

Make sure bootstrap nodes are running before connecting to them. Start bootstrap nodes first, then connect other nodes to them.

### Database Migration Errors

If you encounter database errors, try deleting the database file and letting it recreate:

```bash
rm p2p.db
./bin/p2p serve
```

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]

