## Introduction

The project set out to build a decentralized peer-to-peer storage system that could serve non-technical, collaborative groups while stretching my technical abilities beyond the university curriculum. I approached this work as an opportunity to explore distributed systems, networking, and security concepts that I had not previously implemented in full.

The final artefact is a Go-based application that offers a CLI for storing, retrieving, listing, and deleting encrypted files across a TCP-connected network of peers. It employs a content-addressable storage layout, AES encryption, peer discovery through bootstrap nodes, and an SQLite-backed metadata layer. A demo workflow starts three local nodes, circulates a sample file, and demonstrates end-to-end retrieval, reflecting the system’s current usability.

One major deviation from the proposal emerged in the replication strategy. I originally intended to chunk files and distribute fragments across the network, but the small network size made that approach impractical. Instead, the implemented system encrypts an entire file and replicates the full ciphertext to participating peers, trading storage efficiency for deployability within the available node count.

