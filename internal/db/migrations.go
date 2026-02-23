package db

// migrate runs all schema migrations. Idempotent (uses IF NOT EXISTS).
func (d *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS books (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			author TEXT NOT NULL,
			isbn TEXT,
			genre TEXT DEFAULT '',
			pages INTEGER DEFAULT 0,
			year INTEGER DEFAULT 0,
			cover_path TEXT DEFAULT '',
			cover_source TEXT DEFAULT '',
			status TEXT NOT NULL DEFAULT 'unread',
			rating INTEGER NOT NULL DEFAULT 0,
			notes TEXT DEFAULT '',
			started_at DATETIME,
			finished_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE UNIQUE INDEX IF NOT EXISTS idx_books_isbn ON books(isbn) WHERE isbn IS NOT NULL AND isbn != ''`,

		`CREATE TABLE IF NOT EXISTS shelves (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS book_shelves (
			book_id INTEGER NOT NULL,
			shelf_id INTEGER NOT NULL,
			PRIMARY KEY (book_id, shelf_id),
			FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
			FOREIGN KEY (shelf_id) REFERENCES shelves(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		)`,

		`CREATE TABLE IF NOT EXISTS book_tags (
			book_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			PRIMARY KEY (book_id, tag_id),
			FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)`,

		// FTS5 virtual table for full-text search
		`CREATE VIRTUAL TABLE IF NOT EXISTS books_fts USING fts5(
			title,
			author,
			notes,
			genre,
			content='books',
			content_rowid='id',
			tokenize='porter unicode61'
		)`,

		// Triggers to keep FTS index in sync with books table
		`CREATE TRIGGER IF NOT EXISTS books_ai AFTER INSERT ON books BEGIN
			INSERT INTO books_fts(rowid, title, author, notes, genre)
			VALUES (new.id, new.title, new.author, new.notes, new.genre);
		END`,

		`CREATE TRIGGER IF NOT EXISTS books_ad AFTER DELETE ON books BEGIN
			INSERT INTO books_fts(books_fts, rowid, title, author, notes, genre)
			VALUES ('delete', old.id, old.title, old.author, old.notes, old.genre);
		END`,

		`CREATE TRIGGER IF NOT EXISTS books_au AFTER UPDATE ON books BEGIN
			INSERT INTO books_fts(books_fts, rowid, title, author, notes, genre)
			VALUES ('delete', old.id, old.title, old.author, old.notes, old.genre);
			INSERT INTO books_fts(rowid, title, author, notes, genre)
			VALUES (new.id, new.title, new.author, new.notes, new.genre);
		END`,

		// Loans table for lending and borrowing tracking
		`CREATE TABLE IF NOT EXISTS loans (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			book_id INTEGER NOT NULL,
			loan_type TEXT NOT NULL DEFAULT 'lent',
			person_name TEXT NOT NULL,
			person_contact TEXT DEFAULT '',
			checked_out DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			due_date DATETIME,
			checked_in DATETIME,
			notes TEXT DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
		)`,

		`CREATE INDEX IF NOT EXISTS idx_loans_book_id ON loans(book_id)`,
		`CREATE INDEX IF NOT EXISTS idx_loans_checked_in ON loans(checked_in)`,

		// --- User authentication tables ---

		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE COLLATE NOCASE,
			display_name TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			is_admin INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at)`,

		// Add user_id column to loans (nullable for backwards compatibility with existing data)
		// New loans will have user_id set; old ones will be NULL.
	}

	for _, m := range migrations {
		if _, err := d.Exec(m); err != nil {
			return err
		}
	}

	// Add user_id to loans if not exists (ALTER TABLE IF NOT EXISTS isn't supported in SQLite)
	d.addColumnIfNotExists("loans", "user_id", "INTEGER REFERENCES users(id)")

	return nil
}

// addColumnIfNotExists adds a column to a table if it doesn't already exist.
func (d *DB) addColumnIfNotExists(table, column, colType string) {
	// Check if column exists by querying table_info
	rows, err := d.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, typ string
		var notnull int
		var dflt interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			continue
		}
		if name == column {
			return // column already exists
		}
	}

	d.Exec("ALTER TABLE " + table + " ADD COLUMN " + column + " " + colType)
}
