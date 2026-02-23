package server

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/user/library/internal/models"
	"github.com/user/library/internal/service"
)

type contextKey string

const userContextKey contextKey = "user"

// UserFromContext extracts the authenticated user from the request context.
func UserFromContext(r *http.Request) *models.User {
	if u, ok := r.Context().Value(userContextKey).(*models.User); ok {
		return u
	}
	return nil
}

// auth middleware checks for a valid session cookie or Authorization header.
// Public paths (login, register, static, sw.js) are exempt.
func auth(lib *service.Library) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Public paths that don't require authentication
			if isPublicPath(path) {
				next.ServeHTTP(w, r)
				return
			}

			// Try to get token from cookie first, then Authorization header
			token := ""
			if cookie, err := r.Cookie("session_token"); err == nil {
				token = cookie.Value
			}
			if token == "" {
				auth := r.Header.Get("Authorization")
				if strings.HasPrefix(auth, "Bearer ") {
					token = strings.TrimPrefix(auth, "Bearer ")
				}
			}

			if token == "" {
				handleUnauthorized(w, r)
				return
			}

			user, err := lib.ValidateSession(token)
			if err != nil {
				// Clear invalid cookie
				http.SetCookie(w, &http.Cookie{
					Name:     "session_token",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
				})
				handleUnauthorized(w, r)
				return
			}

			// Inject user into context
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func isPublicPath(path string) bool {
	// Auth pages
	if path == "/login" || path == "/register" {
		return true
	}
	// Auth API endpoints
	if path == "/api/auth/login" || path == "/api/auth/register" {
		return true
	}
	// Static assets, service worker, covers, offline
	if strings.HasPrefix(path, "/static/") || path == "/sw.js" || path == "/offline" {
		return true
	}
	if strings.HasPrefix(path, "/covers/") {
		return true
	}
	return false
}

func handleUnauthorized(w http.ResponseWriter, r *http.Request) {
	// If it's an API request, return JSON error
	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
		return
	}
	// Otherwise redirect to login
	http.Redirect(w, r, "/login", http.StatusFound)
}

// logging wraps an http.Handler to log requests.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(wrapped, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, wrapped.status, time.Since(start).Round(time.Millisecond))
	})
}

// cors adds permissive CORS headers (appropriate for a personal/local server).
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(wrappedWriter(w), r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func wrappedWriter(w http.ResponseWriter) http.ResponseWriter {
	if sw, ok := w.(*statusWriter); ok {
		return sw
	}
	return w
}
