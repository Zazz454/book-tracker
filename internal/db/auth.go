package db

import (
	"database/sql"
	"time"

	"github.com/user/library/internal/models"
)

// --- User queries ---

// CreateUser inserts a new user and returns the ID.
func (d *DB) CreateUser(username, displayName, passwordHash string, isAdmin bool) (int64, error) {
	adminInt := 0
	if isAdmin {
		adminInt = 1
	}
	res, err := d.Exec(
		`INSERT INTO users (username, display_name, password_hash, is_admin) VALUES (?, ?, ?, ?)`,
		username, displayName, passwordHash, adminInt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetUserByUsername retrieves a user by username (case-insensitive).
func (d *DB) GetUserByUsername(username string) (*models.User, error) {
	u := &models.User{}
	var isAdmin int
	err := d.QueryRow(
		`SELECT id, username, display_name, password_hash, is_admin, created_at
		 FROM users WHERE username = ?`, username,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.PasswordHash, &isAdmin, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.IsAdmin = isAdmin != 0
	return u, nil
}

// GetUserByID retrieves a user by ID.
func (d *DB) GetUserByID(id int64) (*models.User, error) {
	u := &models.User{}
	var isAdmin int
	err := d.QueryRow(
		`SELECT id, username, display_name, password_hash, is_admin, created_at
		 FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.PasswordHash, &isAdmin, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.IsAdmin = isAdmin != 0
	return u, nil
}

// ListUsers returns all users.
func (d *DB) ListUsers() ([]models.User, error) {
	rows, err := d.Query(
		`SELECT id, username, display_name, is_admin, created_at FROM users ORDER BY username ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var isAdmin int
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &isAdmin, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.IsAdmin = isAdmin != 0
		users = append(users, u)
	}
	if users == nil {
		users = []models.User{}
	}
	return users, nil
}

// UpdateUserPassword updates a user's password hash.
func (d *DB) UpdateUserPassword(userID int64, passwordHash string) error {
	res, err := d.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, passwordHash, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateUserProfile updates a user's display name.
func (d *DB) UpdateUserProfile(userID int64, displayName string) error {
	res, err := d.Exec(`UPDATE users SET display_name = ? WHERE id = ?`, displayName, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteUser removes a user by ID.
func (d *DB) DeleteUser(id int64) error {
	res, err := d.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CountUsers returns the total number of users.
func (d *DB) CountUsers() (int, error) {
	var count int
	err := d.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// --- Session queries ---

// CreateSession inserts a new session token.
func (d *DB) CreateSession(token string, userID int64, expiresAt time.Time) error {
	_, err := d.Exec(
		`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expiresAt,
	)
	return err
}

// GetSession retrieves a session by token, returning nil if expired or not found.
func (d *DB) GetSession(token string) (*models.Session, error) {
	s := &models.Session{}
	err := d.QueryRow(
		`SELECT token, user_id, expires_at, created_at FROM sessions WHERE token = ?`, token,
	).Scan(&s.Token, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	if time.Now().After(s.ExpiresAt) {
		// Expired - clean it up
		d.DeleteSession(token)
		return nil, sql.ErrNoRows
	}
	return s, nil
}

// DeleteSession removes a session by token.
func (d *DB) DeleteSession(token string) error {
	_, err := d.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

// DeleteUserSessions removes all sessions for a user.
func (d *DB) DeleteUserSessions(userID int64) error {
	_, err := d.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}

// CleanExpiredSessions removes all expired sessions.
func (d *DB) CleanExpiredSessions() error {
	_, err := d.Exec("DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP")
	return err
}
