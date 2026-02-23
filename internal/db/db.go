package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps the sql.DB connection with library-specific helpers.
type DB struct {
	*sql.DB
	dataDir string
}

// Open creates or opens a SQLite database at dataDir/library.db.
// It enables WAL mode and foreign keys, then runs migrations.
func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(dataDir, "covers"), 0755); err != nil {
		return nil, fmt.Errorf("create covers dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "library.db")
	sqlDB, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(wal)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	d := &DB{DB: sqlDB, dataDir: dataDir}

	// Run schema migrations
	if err := d.migrate(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	return d, nil
}

// DataDir returns the data directory path.
func (d *DB) DataDir() string {
	return d.dataDir
}

// CoversDir returns the path to the covers directory.
func (d *DB) CoversDir() string {
	return filepath.Join(d.dataDir, "covers")
}
