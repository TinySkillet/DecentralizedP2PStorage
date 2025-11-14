## Room for Improvement and Future Work

The current artefact demonstrates that a ground-up peer-to-peer storage system can run reliably across a small network, yet it remains a foundation rather than a production deployment. In this section I outline immediate refinements that would polish the existing implementation and describe a longer-term roadmap for expanding its capabilities.

### Part I: Room for Improvement (Refining the Current Artefact)

#### 1. Network Efficiency through Compression
Right now the protocol ships raw encrypted payloads, which means every byte of ciphertext travels uncompressed. Introducing on-the-fly compression—wrapping file streams with `gzip` before encryption—would shrink transfer sizes, conserve bandwidth, and accelerate replication across slow or congested links. Because every store operation broadcasts the full file to multiple peers, compression would yield compounding gains with minimal engineering risk.

#### 2. Enhanced User Experience with a Terminal User Interface (TUI)
The existing CLI is serviceable but austere. Building a terminal user interface using a Go framework such as `Bubble Tea` would transform usability. I envision a dashboard that lists connected peers, surfaces live progress bars for uploads and downloads, and lets users browse files interactively. Presenting real-time state in a single pane would make the system accessible to non-technical collaborators and reduce friction during demonstrations.

#### 3. System Robustness with Graceful Shutdown
Abrupt termination (for example, pressing `Ctrl+C`) currently drops a node without warning its neighbours. Implementing graceful shutdown by trapping signals like `SIGINT` would let the application run a cleanup routine before exit. That routine could broadcast a “leaving” message and flush pending writes, enabling peers to update their address books immediately and preventing ghost entries that cause confusion when the node rejoins.

### Part II: Future Work (Extending the Project’s Capabilities)

#### 1. Secure and Decentralized Peer Discovery
Static bootstrap nodes provide a convenient entry point, but they also introduce a single point of coordination. A secure gossip protocol would allow peers to exchange partial peer lists with random neighbours on a schedule, letting the network discover members organically and heal around failures. Pairing this with cryptographic peer identities—public/private key pairs used to sign announcements—would deter malicious nodes from injecting forged addresses.

#### 2. Architectural Flexibility with Pluggable Transports
The networking layer is currently hard-wired to TCP. Refactoring around an explicit `Transport` interface would decouple application logic from the underlying protocol and open the door to alternative transports. Future implementations could include gRPC for schema-driven RPCs, WebSockets to invite browser clients into the swarm, or QUIC to improve performance over lossy links. This modularity would make the system adaptable to varied deployment environments.

#### 3. Advanced Redundancy and Data Healing
Full replication offers simplicity, yet it scales poorly as datasets grow. Adding erasure coding as a user-selectable option would deliver storage-efficient redundancy while preserving availability. Complementing that with automated data healing—a background process that audits replica counts, detects missing shards, and re-replicates from healthy peers—would bring the system closer to production-grade resilience.

#### 4. Access Control and Multi-Tenancy
The current design assumes trusted peers and a shared encryption key. Introducing an access-control layer anchored in cryptographic identities would allow multiple groups to share infrastructure without sharing data. Group-specific encryption keys and policy enforcement could ensure that only authorised members decrypt particular datasets, enabling genuine multi-tenancy while retaining the privacy guarantees that motivated the project.

