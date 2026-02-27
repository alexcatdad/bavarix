package vault

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var ErrNoBackup = errors.New("vault: no backup found")

type Entry struct {
	ID        int64
	Chassis   string
	Module    string
	Version   string
	Data      []byte
	CreatedAt time.Time
}

type Vault struct {
	db *sql.DB
}

func Open(path string) (*Vault, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("vault: opening database: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS backups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chassis TEXT NOT NULL,
			module TEXT NOT NULL,
			version TEXT NOT NULL,
			data BLOB NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("vault: creating table: %w", err)
	}

	return &Vault{db: db}, nil
}

func (v *Vault) Close() error {
	return v.db.Close()
}

func (v *Vault) Save(chassis, module, version string, data []byte) error {
	_, err := v.db.Exec(
		`INSERT INTO backups (chassis, module, version, data) VALUES (?, ?, ?, ?)`,
		chassis, module, version, data,
	)
	if err != nil {
		return fmt.Errorf("vault: saving backup: %w", err)
	}
	return nil
}

func (v *Vault) List(chassis, module string) ([]Entry, error) {
	rows, err := v.db.Query(
		`SELECT id, chassis, module, version, data, created_at
		 FROM backups
		 WHERE chassis = ? AND module = ?
		 ORDER BY id DESC`,
		chassis, module,
	)
	if err != nil {
		return nil, fmt.Errorf("vault: listing backups: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Chassis, &e.Module, &e.Version, &e.Data, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("vault: scanning row: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (v *Vault) Latest(chassis, module string) (Entry, error) {
	entries, err := v.List(chassis, module)
	if err != nil {
		return Entry{}, err
	}
	if len(entries) == 0 {
		return Entry{}, ErrNoBackup
	}
	return entries[0], nil
}
