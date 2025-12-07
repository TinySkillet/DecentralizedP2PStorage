#!/bin/bash

# Test script for peer gossip protocol
# This demonstrates automatic peer discovery through gossip

set -e

echo "=== Peer Gossip Protocol Test ==="
echo ""
echo "This test will:"
echo "1. Start Node A on :3000"
echo "2. Start Node B on :4000, bootstrapping to Node A"
echo "3. Start Node C on :5000, bootstrapping ONLY to Node B"
echo "4. Node C should automatically discover and connect to Node A!"
echo ""

# Build if needed
if [ ! -f "./DecentralizedP2PStorage" ]; then
    echo "Building application..."
    go build
fi

# Cleanup old data
echo "Cleaning up old test data..."
rm -rf node_3000 node_4000 node_5000

# Create directories for logs and databases
echo "Creating node directories..."
mkdir -p node_3000 node_4000 node_5000

echo ""
echo "Starting nodes..."
echo ""

# Start Node A in background
echo "[Node A] Starting on :3000..."
./DecentralizedP2PStorage serve --listen :3000 --db node_3000/p2p.db > node_3000/node_a.log 2>&1 &
PID_A=$!
sleep 2

# Start Node B in background
echo "[Node B] Starting on :4000, bootstrapping to :3000..."
./DecentralizedP2PStorage serve --listen :4000 --db node_4000/p2p.db --bootstrap localhost:3000 > node_4000/node_b.log 2>&1 &
PID_B=$!
sleep 3

# Start Node C in foreground (we want to see its discovery logs)
echo "[Node C] Starting on :5000, bootstrapping to :4000..."
echo ""
echo "====================================="
echo "WATCH FOR PEER DISCOVERY IN LOGS:"
echo "====================================="
echo ""
./DecentralizedP2PStorage serve --listen :5000 --db node_5000/p2p.db --bootstrap localhost:4000 &
PID_C=$!

# Wait a bit for connections to establish
sleep 5

echo ""
echo "====================================="
echo "VERIFICATION"
echo "====================================="
echo ""

# Check Node C's database
echo "Node C's peer list (should include both :3000 and :4000):"
sqlite3 node_5000/p2p.db "SELECT address, status, datetime(last_seen, 'localtime') as last_seen FROM peers ORDER BY last_seen DESC;" 2>/dev/null || echo "Database not ready yet"

echo ""
echo "Node C should have discovered Node A through Node B!"
echo ""
echo "Press Ctrl+C to stop all nodes..."
echo ""

# Wait for user to stop
wait $PID_C

# Cleanup
echo ""
echo "Stopping all nodes..."
kill $PID_A $PID_B $PID_C 2>/dev/null || true

echo ""
echo "=== Test Complete ==="
echo ""
echo "Check the logs:"
echo "  - node_3000/node_a.log"
echo "  - node_4000/node_b.log"
echo "  - Node C logs (displayed above)"
echo ""
echo "Check peer databases:"
echo "  sqlite3 node_3000/p2p.db 'SELECT * FROM peers;'"
echo "  sqlite3 node_4000/p2p.db 'SELECT * FROM peers;'"
echo "  sqlite3 node_5000/p2p.db 'SELECT * FROM peers;'"
