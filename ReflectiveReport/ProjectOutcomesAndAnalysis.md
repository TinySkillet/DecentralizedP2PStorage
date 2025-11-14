## Analysis of Progress Made

### 1. Introduction: Overall Project Outcome

The project was successful and resulted in a functional decentralized storage application. The final artefact enabled me to store, retrieve, list, and delete encrypted files across a peer-to-peer network of nodes communicating over TCP. While the core vision—privacy-preserving storage with user control for small groups—remained intact, several implementation details evolved to match the realities of a small network and the project’s educational goals.

### 2. Mapping Achievements Against Original Objectives

#### Objective 1: Build a secure P2P network
This objective was met through a custom TCP-based networking stack with a clear handshake, message framing, and peer coordination. Nodes discovered peers via bootstrap nodes, allowing new participants to join a running network without a public DHT. For confidentiality, files were encrypted using AES (AES-256 in practice) before replication and storage on untrusted peers. Together, these design choices achieved an end-to-end secure-by-default flow appropriate for a small, cooperative peer set.

Evidence in the artefact:
- Custom TCP transport with length-prefixed message framing and a simple handshake protocol.
- Bootstrap-based peer discovery for initial connectivity and peer list exchange.
- AES-based file encryption prior to storage and transfer.

#### Objective 2: Implement a content-addressable storage (CAS) layer
This objective was met by implementing a SHA-256–based CAS scheme. Files were addressed by the hash of their content, and the on-disk layout derived directories from hash prefixes. This ensured self-verifying data and straightforward deduplication: identical content maps to the same address.

Evidence in the artefact:
- SHA-256 hashing for addressing and verification.
- Directory layout keyed by hash prefixes to avoid hot spots and support scalable organization.
- Retrieval logic that validates content against expected digests.

#### Objective 3: Provide redundancy for availability
This objective was met via full replication. Every participating peer maintained a complete encrypted copy of each stored file, ensuring maximum availability in the face of peer churn within a small network. Although storage-inefficient at scale, this approach fulfilled the objective’s intent—data remained accessible even if individual nodes went offline.

Evidence in the artefact:
- Store operation broadcasted encrypted content to connected peers.
- Peers served reads for any known key, enabling retrieval independent of a specific node.

#### Objective 4: Deliver a simple CLI for non-technical users
This objective was met with a clear command-line interface that exposed intuitive operations:
- `serve` to run a node (with `--listen` and `--bootstrap` flags)
- `store <key> <file>` to add an encrypted file to the network
- `get <key>` (with `--out`) to retrieve files
- `delete <key>` to remove local copies
- `files list` to inspect known files
- `demo` to spin up a local three-node network for end-to-end testing

### 3. Analysis of Key Deviations and Justifications

The most significant deviation from the proposal was the shift from a chunking-based distribution model to full replication. The initial idea was to break files into chunks and distribute fragments across peers, inspired by large-scale systems where storage efficiency and parallel retrieval matter. In practice, for a small and private network, chunking introduced unnecessary complexity and brittleness: coordinating fragment placement, reassembly logic, and recovery semantics demanded additional protocols that would not materially improve reliability at small node counts. Choosing full replication simplified both the implementation and the operational model, improved robustness, and delivered predictable performance. This pivot was a pragmatic and well-reasoned engineering decision that prioritized reliability and clarity over premature optimization—precisely the right trade-off for the target audience.

### 4. Reflection on Development Methodology and Process

The development process was iterative and empirical. I started with the minimum viable transport—establishing a TCP connection between two peers—then layered in essentials: a handshake, length-prefixed message framing, serialization for distinct message types, and basic peer discovery via bootstrap nodes. On top of that substrate, I added the CAS layer (hashing and storage layout), then encryption, and finally the CLI to expose end-user workflows. Each step was validated before moving to the next, allowing issues at the transport or protocol boundary to be discovered early.

For tooling, I used Git for version control to manage changes and branching, and Visual Studio Code as the primary IDE. My testing strategy focused on manual integration tests that simulated realistic usage. I routinely ran multiple local peer instances in separate terminals—commonly on `:3000`, `:4000`, and `:5000`—to form a small network. I then exercised scenarios such as storing a file from one node, retrieving it from another, deleting local copies, and verifying that retrieval still worked via remaining peers. I also tested peer churn by starting and stopping nodes while ensuring data availability held under replication. This hands-on approach gave rapid feedback about framing errors, handshake edge cases, and correctness of the CAS and encryption paths.

In sum, the project not only met its objectives but did so in a way that strengthened my understanding of distributed systems fundamentals. The deviations made were deliberate and aligned with the project’s constraints and goals, resulting in an artefact that is both functional and educationally rich.


(Include Visuals like screenshots of CLI, architecture diagram)