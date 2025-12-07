# Flowchart Prompt for Eraser.io

## Decentralized P2P Storage System - Operation Flowcharts

Create flowcharts showing the sequence of operations in a peer-to-peer file storage system:

### Store Operation Flowchart

**Start:** User calls `store <key> <file>`

1. **FileServer receives file**
2. **Compute SHA-256 hash** of file content
3. **Check if content hash exists** in Storage
   - **If Yes:** Skip storage (deduplication), go to step 5
   - **If No:** Continue to step 4
4. **Write file to Storage** using content hash as address
5. **Store metadata in Database** (user key → content hash mapping)
6. **Encrypt file** with AES-256
7. **Broadcast MessageStoreFile** {hash, size} to all connected peers
8. **Stream encrypted file data** to all peers
9. **For each receiving peer:**
   - Receive MessageStoreFile
   - Receive encrypted stream
   - Decrypt file
   - Check if content hash exists in Storage
     - **If Yes:** Skip (deduplication)
     - **If No:** Write to Storage using content hash
10. **End:** File stored on all peers

### Get Operation Flowchart

**Start:** User calls `get <key>`

1. **FileServer receives request**
2. **Query Database** for content hash (key → hash lookup)
3. **Check if file exists locally** in Storage using content hash
   - **If Yes:** Read from Storage → Return to user → **End**
   - **If No:** Continue to step 4
4. **Broadcast MessageGetFile** {contentHash} to all connected peers
5. **Wait for response** from peers
6. **For each peer:**
   - Peer receives MessageGetFile
   - Peer checks Storage for content hash
     - **If Found:** Stream file data to requester
     - **If Not Found:** Ignore request
7. **Requester receives encrypted stream**
8. **Decrypt file**
9. **Write to local Storage** using content hash
10. **Return file to user**
11. **End**

### Delete Operation Flowchart

**Start:** User calls `delete <key>`

1. **FileServer receives request**
2. **Query Database** for content hash (key → hash lookup)
3. **Delete from local Storage** using content hash
4. **Delete metadata from Database**
5. **Broadcast MessageDeleteFile** {contentHash} to all connected peers
6. **For each receiving peer:**
   - Receive MessageDeleteFile
   - Query Database for all files with this content hash
   - Delete from Storage using content hash
   - Delete metadata from Database
7. **End:** File deleted from all peers

### Visual Style
- Use standard flowchart shapes:
  - Rectangles for processes
  - Diamonds for decisions (Yes/No branches)
  - Arrows for flow direction
  - Parallelograms for input/output
- Label decision points clearly
- Show message exchanges between nodes
- Keep flows readable and sequential
- Use different colors for different operations (optional)

