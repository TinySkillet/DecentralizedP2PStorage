package db

import (
	"context"
	"time"
)

type Key struct {
	ID        string
	Label     string
	Algo      string
	KeyBytes  []byte
	CreatedAt time.Time
}

type File struct {
	ID        string
	Name      string
	Hash      string
	Size      int64
	LocalPath string
	CreatedAt time.Time
}

type Peer struct {
	ID       string
	Address  string
	Status   string
	LastSeen *time.Time
}

type Share struct {
	ID        string
	FileID    string
	PeerID    string
	Direction string
	CreatedAt time.Time
}

func (d *DB) UpsertPeer(ctx context.Context, p Peer) error {
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO peers(id,address,status,last_seen)
		VALUES(?,?,?,?)
		ON CONFLICT(address) DO UPDATE SET
			status=excluded.status,
			last_seen=excluded.last_seen
	`, p.ID, p.Address, p.Status, p.LastSeen)
	return err
}

func (d *DB) InsertFileWithKey(ctx context.Context, f File, keyID string) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO files(id,name,hash,size,local_path)
		VALUES(?,?,?,?,?)
	`, f.ID, f.Name, f.Hash, f.Size, f.LocalPath); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT OR IGNORE INTO file_keys(file_id,key_id)
		VALUES(?,?)
	`, f.ID, keyID); err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) ListFiles(ctx context.Context) ([]File, error) {
	rows, err := d.sql.QueryContext(ctx, `
		SELECT id,name,hash,size,local_path,created_at FROM files ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []File
	for rows.Next() {
		var f File
		if err := rows.Scan(&f.ID, &f.Name, &f.Hash, &f.Size, &f.LocalPath, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// GetKey returns a key by id.
func (d *DB) GetKey(ctx context.Context, id string) (*Key, error) {
	row := d.sql.QueryRowContext(ctx, `
		SELECT id,label,algo,key_bytes,created_at FROM keys WHERE id=?
	`, id)
	var k Key
	if err := row.Scan(&k.ID, &k.Label, &k.Algo, &k.KeyBytes, &k.CreatedAt); err != nil {
		return nil, err
	}
	return &k, nil
}

// PutKey inserts or replaces a key.
func (d *DB) PutKey(ctx context.Context, k Key) error {
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO keys(id,label,algo,key_bytes,created_at)
		VALUES(?,?,?,?,CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			label=excluded.label,
			algo=excluded.algo,
			key_bytes=excluded.key_bytes
	`, k.ID, k.Label, k.Algo, k.KeyBytes)
	return err
}

// GetOrCreateDefaultKey returns bytes for key id "default"; creates it if missing.
func (d *DB) GetOrCreateDefaultKey(ctx context.Context, gen func() []byte) ([]byte, error) {
	const id = "default"
	k, err := d.GetKey(ctx, id)
	if err == nil {
		return k.KeyBytes, nil
	}
	// If not found, create.
	keyBytes := gen()
	if err := d.PutKey(ctx, Key{
		ID:       id,
		Label:    "default",
		Algo:     "AES-CTR-256",
		KeyBytes: keyBytes,
	}); err != nil {
		return nil, err
	}
	return keyBytes, nil
}
