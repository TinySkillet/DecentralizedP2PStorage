## Project Timeline and Management

### 1. Introduction: A Learning-Driven Approach

I began the project in July with an intentionally organic plan that put learning first. Rather than sprinting into coding, I structured the work around four phases—Foundational Learning, Architectural Design, Phased Implementation, and Finalization—so that each stage could build on the skills developed in the one before it.

### 2. Phase 1: Foundational Skill Acquisition (July)

The first month was almost entirely devoted to study. I immersed myself in online resources—YouTube walkthroughs, technical blogs, and documentation—to understand how Go handles TCP networking. I experimented with sample programs to see how sockets behave, how keep-alive messages prevent idle connections from dying, and how channels coordinate concurrent goroutines. That time investment was crucial. By the end of July, I had the confidence to sketch and eventually own the full networking layer instead of leaning on a third-party abstraction.

### 3. Phase 2: Architectural Design and Planning (August)

Armed with those fundamentals, August became the design month. I spent this phase brainstorming the overall architecture and translating ideas into a concrete technical blueprint. I formalised the commitment to a custom CAS layer, bootstrap-based peer discovery, and a full replication strategy that matched the project’s scale. The prior learning made these decisions realistic: I knew exactly what I was asking myself to build and which trade-offs would keep the system approachable without diluting its decentralised character. Planning felt less like speculation and more like a roadmap grounded in the behaviours I had already explored.

### 4. Phase 3: Phased Implementation (September – October)

The bulk of development happened across September and October. I worked in layers to manage complexity. First, I implemented the networking core: the TCP transport, message framing, and peer lifecycle. Next came the storage layer, turning SHA-256 hashes into directory paths and streaming data straight to disk. Integration followed, wiring transport events into the file server and ensuring messages triggered the right storage operations. Finally, I layered on AES encryption and expanded the CLI so users could run nodes, store files, and retrieve them without touching internal APIs. This iterative approach let me validate each component before introducing the next, catching bugs where they originated instead of hunting for them later.

### 5. Phase 4: Finalization and Testing (November)

November focused on polish and assurance. I relied on manual integration tests—running three local nodes in separate terminals, simulating peer churn, and observing data replication—to confirm that the system behaved under realistic conditions. I refined command ergonomics, adding flags and help text, and tidied logs so network activity remained legible. This closing phase reinforced how essential hands-on testing is when formal test harnesses lag behind; it helped me ship an artefact that felt dependable, not just functional.

