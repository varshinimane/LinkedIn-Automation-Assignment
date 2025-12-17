package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Init opens sqlite database and ensures schema.
func Init(path string) (*sql.DB, error) {
	// Ensure parent directory exists so sqlite can create the file.
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := ensureSchema(db); err != nil {
		return nil, err
	}
	return db, nil
}

func ensureSchema(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS connections (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	profile_url TEXT UNIQUE,
	name TEXT,
	company TEXT,
	status TEXT,
	contacted_at DATETIME,
	accepted_at DATETIME,
	last_message_at DATETIME
);
CREATE TABLE IF NOT EXISTS messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	profile_url TEXT,
	body TEXT,
	sent_at DATETIME
);
CREATE TABLE IF NOT EXISTS counters (
	key TEXT PRIMARY KEY,
	value INTEGER
);
`
	_, err := db.Exec(schema)
	return err
}
