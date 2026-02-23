package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/user/library/internal/models"
)

const (
	sessionDuration = 30 * 24 * time.Hour // 30 days
	bcryptCost      = 12
	minPasswordLen  = 4
	minUsernameLen  = 2
)

// Register creates a new user account.
func (l *Library) Register(req *models.RegisterRequest, isAdmin bool) (*models.User, error) {
	username := strings.TrimSpace(strings.ToLower(req.Username))
	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = req.Username
	}

	if len(username) < minUsernameLen {
		return nil, fmt.Errorf("username must be at least %d characters", minUsernameLen)
	}
	if len(req.Password) < minPasswordLen {
		return nil, fmt.Errorf("password must be at least %d characters", minPasswordLen)
	}

	// Check if username is taken
	existing, _ := l.DB.GetUserByUsername(username)
	if existing != nil {
		return nil, fmt.Errorf("username %q is already taken", username)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	id, err := l.DB.CreateUser(username, displayName, string(hash), isAdmin)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return l.DB.GetUserByID(id)
}

// Login validates credentials and creates a session.
func (l *Library) Login(username, password string) (*models.LoginResponse, error) {
	user, err := l.DB.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	expiresAt := time.Now().Add(sessionDuration)
	if err := l.DB.CreateSession(token, user.ID, expiresAt); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	// Clear password hash before returning
	user.PasswordHash = ""

	return &models.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// Logout deletes a session token.
func (l *Library) Logout(token string) error {
	return l.DB.DeleteSession(token)
}

// ValidateSession checks if a session token is valid and returns the user.
func (l *Library) ValidateSession(token string) (*models.User, error) {
	session, err := l.DB.GetSession(token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	user, err := l.DB.GetUserByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	user.PasswordHash = ""
	return user, nil
}

// ChangePassword changes a user's password.
func (l *Library) ChangePassword(userID int64, currentPassword, newPassword string) error {
	user, err := l.DB.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	if len(newPassword) < minPasswordLen {
		return fmt.Errorf("new password must be at least %d characters", minPasswordLen)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return l.DB.UpdateUserPassword(userID, string(hash))
}

// UpdateProfile updates a user's display name.
func (l *Library) UpdateProfile(userID int64, displayName string) (*models.User, error) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return nil, fmt.Errorf("display name is required")
	}
	if err := l.DB.UpdateUserProfile(userID, displayName); err != nil {
		return nil, err
	}
	return l.DB.GetUserByID(userID)
}

// ListUsers returns all users (admin only).
func (l *Library) ListUsers() ([]models.User, error) {
	return l.DB.ListUsers()
}

// DeleteUser removes a user (admin only).
func (l *Library) DeleteUser(id int64) error {
	// Delete their sessions first
	l.DB.DeleteUserSessions(id)
	return l.DB.DeleteUser(id)
}

// CountUsers returns the total number of registered users.
func (l *Library) CountUsers() (int, error) {
	return l.DB.CountUsers()
}

// EnsureDefaultAdmin creates a default admin account if no users exist.
// Returns true if a new admin was created.
func (l *Library) EnsureDefaultAdmin(password string) (bool, error) {
	count, err := l.DB.CountUsers()
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}

	_, err = l.Register(&models.RegisterRequest{
		Username:    "admin",
		DisplayName: "Admin",
		Password:    password,
	}, true)
	if err != nil {
		return false, err
	}
	return true, nil
}

// IsRegistrationOpen returns true if new users can self-register.
// Currently: allow registration if no users exist (first user becomes admin),
// otherwise only admins can create users.
func (l *Library) IsRegistrationOpen() bool {
	count, _ := l.DB.CountUsers()
	return count == 0
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
