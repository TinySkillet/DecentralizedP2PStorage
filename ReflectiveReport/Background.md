## Background

I set out to build a decentralized peer-to-peer storage system that could give small, collaborative groups control over their data without depending on a centralized provider. The goal was to provide a privacy-focused alternative to services like Google Drive, where trust in a third party and vendor lock-in remain constant concerns.

Before the project, my technical understanding was mostly theoretical. I knew the principles of peer-to-peer networking, but I had never implemented peer discovery or managed live connections between nodes. Content-addressable storage made sense conceptually from tools like Git, yet I had not designed a storage layer around hashes in practice. I also understood how AES encryption works, but I had never applied it to protect data stored on machines that I did not trust.

Ensuring data availability became the central design trade-off. I initially planned to chunk files and scatter pieces across the network, but that approach proved too complex and brittle for a small number of peers. I ultimately chose full replication: every node stores the complete encrypted file. This decision favoured simplicity and high availability over storage efficiency, which matches the needs of tight-knit groups that prioritise reliability and control over raw capacity.

