#!/bin/bash
# Setup script for local development nodes

echo "Setting up local node directories..."

# Create directories for 3 nodes
mkdir -p node_3000
mkdir -p node_4000
mkdir -p node_5000

echo "Created directories:"
echo "  - node_3000"
echo "  - node_4000"
echo "  - node_5000"
echo ""
echo "You can now run the nodes:"
echo "  1. ./bin/p2p serve --listen :3000 --db node_3000/p2p.db"
echo "  2. ./bin/p2p serve --listen :4000 --db node_4000/p2p.db --bootstrap localhost:3000"
echo "  3. ./bin/p2p serve --listen :5000 --db node_5000/p2p.db --bootstrap localhost:4000"
