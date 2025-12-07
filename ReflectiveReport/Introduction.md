## Introduction

This project's primary goal was to develop a decentralized peer-to-peer storage system tailored for non-technical, collaborative groups, serving as a practical application to extend my technical abilities beyond the university curriculum. I approached this work as an opportunity to implement and explore core concepts in distributed systems, networking, and security; concepts that I had not previously implemented in full.

The final artefact is a Go-based application that offers a Command Line Interface (CLI) for storing, retrieving, listing, and deleting encrypted files across a TCP-connected network of peers. It employs a content-addressable storage layout, AES encryption, peer discovery through bootstrap nodes, and an SQLite-backed metadata layer. 

A significant deviation from the original proposal was the replication strategy. I originally planned to chunk files and distribute fragments across the network, but the small network size made that approach impractical. Instead, the implemented system encrypts an entire file and replicates the full ciphertext to participating peers, trading storage efficiency for deployability within the available node count.

