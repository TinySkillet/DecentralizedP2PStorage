## 1. Project Overview

This is a decentralized peer-to-peer (P2P) file storage system written in Go. The system allows nodes (peers) to connect, share, and retrieve files from each other over a TCP network.

The core architecture consists of:
- A `FileServer` that manages peers, storage, and message handling.
- A `Transport` layer (currently TCP) to handle low-level network communication.
- A `Store` component for interacting with the file system.
- A P2P message-passing system using `gob` for encoding/decoding.

## 2. Core Principles

- **Simplicity and Readability**: Code should be easy to understand. Prioritize clarity over overly clever solutions.
- **Concurrency**: The system is highly concurrent. Pay close attention to race conditions and use mutexes and channels appropriately.
- **Modularity**: Keep components decoupled. The transport layer should not know about the storage layer's implementation details.
- **Error Handling**: Handle all errors gracefully. Do not ignore them. Propagate errors up or log them appropriately.

## 3. Do's and Don'ts

### Do:
- **Use Go Modules**: Manage dependencies with `go mod`. Run `go mod tidy` after adding new dependencies.
- **Follow Standard Go Formatting**: Run `gofmt -w .` or `goimports -w .` on any changed files.
- **Handle Errors Explicitly**: Always check for errors. Use `if err != nil` blocks. For network errors, check for `io.EOF` or `net.ErrClosed` to handle graceful disconnections.
- **Use Channels for Goroutine Communication**: Use channels for signaling and passing data between goroutines (e.g., the `quitch` channel for shutdowns, `rpcChan` for incoming messages).
- **Lock Shared State**: Use `sync.Mutex` to protect access to shared data structures like the `peers` map in `FileServer`.
- **Use Interfaces**: Adhere to the `Peer` and `Transport` interfaces to keep the network layer abstract.
- **Write Unit Tests**: For any new functionality, add corresponding unit tests. Place them in `_test.go` files.

### Don't:
- **Do not use `panic()`**: Avoid using `panic()` for recoverable errors. Return an `error` instead.
- **Do not ignore errors**: Never use `_` to discard an error unless you are certain it's not relevant.
- **Do not introduce global state**: Avoid global variables. Pass dependencies explicitly.
- **Do not block the main event loop**: In `FileServer.loop()`, message handling should be non-blocking. Offload long-running tasks to new goroutines if necessary.
- **Do not use `time.Sleep` for synchronization**: Avoid `time.Sleep` to wait for network responses or goroutines. Use channels or waitgroups for proper synchronization.

## 4. Development Workflow & Commands

- **To run the project locally**:
  ```bash
  go run .
  ```

- **To run tests**:
  ```bash
  go test ./...
  ```

- **To manage dependencies**:
  ```bash
  go mod tidy
  ```

## 5. Project Structure

- `main.go`: Entry point. Responsible for initializing and starting the `FileServer` nodes.
- `server.go`: Contains the `FileServer` logic, which orchestrates all the main components like peer management, message handling, and file storage/retrieval operations.
- `storage.go`: Handles all file system interactions: reading, writing, and deleting files. The `CASPathTransformFunc` creates a content-addressable storage structure.
- `p2p/transport.go`: Defines the `Peer` and `Transport` interfaces, which are the core abstractions for network communication.
- `p2p/tcp_transport.go`: An implementation of the `Transport` interface using TCP. It manages connections and the main event loop for accepting new peers.
- `p2p/message.go`: Defines the `RPC` struct used for communication between peers.
- `p2p/encoding.go`: Defines the `Decoder` interface and its default implementation for reading `RPC` data from a connection.
- `p2p/handshake.go`: Defines the handshake logic for new connections (currently a no-op).

