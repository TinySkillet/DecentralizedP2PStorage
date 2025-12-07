# P2P Storage - Deployment Guide

## Quick Start

### Installation

```bash
./install.sh
```

This will:
- Build the binary
- Install to `/usr/local/bin/`
- Create config directory `~/.p2p/`
- Install systemd service

### Running as a Service

```bash
# Start the service
sudo systemctl start p2p-storage@$USER

# Enable on boot
sudo systemctl enable p2p-storage@$USER

# Check status
sudo systemctl status p2p-storage@$USER

# View logs
sudo journalctl -u p2p-storage@$USER -f
```

---

## Manual Running

### Start a Node

```bash
./DecentralizedP2PStorage serve --listen :3000 --db ~/.p2p/p2p.db
```

### With Bootstrap Nodes

```bash
./DecentralizedP2PStorage serve --listen :4000 --db ~/.p2p/p2p.db --bootstrap localhost:3000
```

---

## Commands

### View Connected Peers

```bash
./DecentralizedP2PStorage peers --db ~/.p2p/p2p.db
```

### Clean Up Stale Peers

```bash
./DecentralizedP2PStorage cleanup --db ~/.p2p/p2p.db
```

### Store a File

```bash
./DecentralizedP2PStorage store myfile document.pdf --listen :7000 --db /tmp/temp.db --bootstrap localhost:3000
```

### Retrieve a File

```bash
./DecentralizedP2PStorage get myfile --listen :7000 --db /tmp/temp.db --bootstrap localhost:3000 --out retrieved.pdf
```

### List Stored Files

```bash
./DecentralizedP2PStorage files list --db ~/.p2p/p2p.db
```

### View File Shares

```bash
./DecentralizedP2PStorage shares --db ~/.p2p/p2p.db
```

---

## Configuration

### Systemd Service

Edit `/etc/systemd/system/p2p-storage@.service` to customize:

- **Listen address:** Change `--listen :3000`
- **Bootstrap nodes:** Change `--bootstrap localhost:4000`
- **Database path:** Change `--db` path

After editing:
```bash
sudo systemctl daemon-reload
sudo systemctl restart p2p-storage@$USER
```

---

## Network Setup

### Firewall Configuration

Allow incoming connections on your listen port:

```bash
# UFW
sudo ufw allow 3000/tcp

# firewalld
sudo firewall-cmd --permanent --add-port=3000/tcp
sudo firewall-cmd --reload
```

### Bootstrap Nodes

For a production network, set up dedicated bootstrap nodes:

1. **Choose 2-3 stable machines** as bootstrap nodes
2. **Start them first** without bootstrap flags
3. **Configure other nodes** to bootstrap to these stable nodes

Example:
```bash
# Bootstrap Node 1 (always-on server)
./DecentralizedP2PStorage serve --listen :3000 --db ~/.p2p/p2p.db

# Bootstrap Node 2 (always-on server)
./DecentralizedP2PStorage serve --listen :3000 --db ~/.p2p/p2p.db --bootstrap bootstrap1.example.com:3000

# Regular nodes
./DecentralizedP2PStorage serve --listen :3000 --db ~/.p2p/p2p.db --bootstrap bootstrap1.example.com:3000,bootstrap2.example.com:3000
```

---

## Monitoring

### Check Number of Peers

```bash
./DecentralizedP2PStorage peers --db ~/.p2p/p2p.db | wc -l
```

### View Logs

```bash
# Systemd service
sudo journalctl -u p2p-storage@$USER -f

# Or grep for specific events
sudo journalctl -u p2p-storage@$USER | grep "Connected with remote"
```

### Database Inspection

```bash
# View all tables
sqlite3 ~/.p2p/p2p.db ".tables"

# Count peers
sqlite3 ~/.p2p/p2p.db "SELECT COUNT(*) FROM peers;"

# View files
sqlite3 ~/.p2p/p2p.db "SELECT name, size FROM files;"
```

---

## Troubleshooting

### Port Already in Use

Change the listen port:
```bash
./DecentralizedP2PStorage serve --listen :3001 --db ~/.p2p/p2p.db
```

### Can't Connect  to Remote Peers

1. Check firewall allows the port
2. Verify IP address is correct
3. Ensure remote node is running

### Database Locked Errors

Only one instance can use a database at a time. Use different DB paths for different nodes:
```bash
./DecentralizedP2PStorage serve --listen :3000 --db ~/.p2p/node1.db
./DecentralizedP2PStorage serve --listen :4000 --db ~/.p2p/node2.db
```

---

## Production Checklist

- [ ] Bootstrap nodes configured and running
- [ ] Firewall ports opened
- [ ] Systemd service enabled
- [ ] Monitoring set up (logs, peer count)
- [ ] Backup strategy for database
- [ ] Regular cleanup of stale peers scheduled
