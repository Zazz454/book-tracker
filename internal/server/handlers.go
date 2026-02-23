package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/user/library/internal/models"
	"github.com/user/library/internal/service"
)

// apiHandlers holds the JSON API handlers.
type apiHandlers struct {
	lib *service.Library
}

func newAPIHandlers(lib *service.Library) *apiHandlers {
	return &apiHandlers{lib: lib}
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseID(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, p := range parts {
		// Look for numeric segments
		if id, err := strconv.ParseInt(p, 10, 64); err == nil && id > 0 {
			_ = i
			return id, nil
		}
	}
	return 0, fmt.Errorf("no valid ID in path: %s", path)
}

// extractPathSegments splits a URL path into segments.
// e.g., "/api/books/5/status" -> ["api", "books", "5", "status"]
func extractPathSegments(path string) []string {
	var segments []string
	for _, s := range strings.Split(path, "/") {
		if s != "" {
			segments = append(segments, s)
		}
	}
	return segments
}

// --- Book handlers ---

func (h *apiHandlers) listBooks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	params := models.BookListParams{
		Status:  q.Get("status"),
		Genre:   q.Get("genre"),
		TagName: q.Get("tag"),
		Sort:    q.Get("sort"),
		Order:   q.Get("order"),
		Query:   q.Get("q"),
	}
	if sid := q.Get("shelf_id"); sid != "" {
		params.ShelfID, _ = strconv.ParseInt(sid, 10, 64)
	}
	if lim := q.Get("limit"); lim != "" {
		params.Limit, _ = strconv.Atoi(lim)
	}
	if off := q.Get("offset"); off != "" {
		params.Offset, _ = strconv.Atoi(off)
	}

	result, err := h.lib.ListBooks(params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *apiHandlers) createBook(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	book, err := h.lib.CreateBook(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, book)
}

func (h *apiHandlers) getBook(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	book, err := h.lib.GetBook(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "book not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (h *apiHandlers) updateBook(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req models.UpdateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	book, err := h.lib.UpdateBook(id, &req)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "book not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (h *apiHandlers) deleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.lib.DeleteBook(id); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "book not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *apiHandlers) updateBookStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	book, err := h.lib.UpdateStatus(id, req.Status)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (h *apiHandlers) updateBookRating(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req models.UpdateRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	book, err := h.lib.UpdateRating(id, req.Rating)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (h *apiHandlers) setCover(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req models.SetCoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	book, err := h.lib.SetCoverURL(id, req.URL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (h *apiHandlers) refreshCover(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	book, err := h.lib.RefreshCover(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (h *apiHandlers) searchBooks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	result, err := h.lib.SearchBooks(q, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// --- Shelf handlers ---

func (h *apiHandlers) listShelves(w http.ResponseWriter, r *http.Request) {
	shelves, err := h.lib.ListShelves()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, shelves)
}

func (h *apiHandlers) createShelf(w http.ResponseWriter, r *http.Request) {
	var req models.CreateShelfRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	shelf, err := h.lib.CreateShelf(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, shelf)
}

func (h *apiHandlers) getShelf(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	shelf, err := h.lib.GetShelf(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "shelf not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, shelf)
}

func (h *apiHandlers) updateShelf(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req models.UpdateShelfRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	shelf, err := h.lib.UpdateShelf(id, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, shelf)
}

func (h *apiHandlers) deleteShelf(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.lib.DeleteShelf(id); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "shelf not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *apiHandlers) addBookToShelf(w http.ResponseWriter, r *http.Request) {
	segs := extractPathSegments(r.URL.Path)
	// /api/shelves/{id}/books
	var shelfID int64
	for i, s := range segs {
		if s == "shelves" && i+1 < len(segs) {
			shelfID, _ = strconv.ParseInt(segs[i+1], 10, 64)
			break
		}
	}
	if shelfID == 0 {
		writeError(w, http.StatusBadRequest, "invalid shelf ID")
		return
	}

	var req models.ShelfBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := h.lib.AddBookToShelf(shelfID, req.BookID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "added"})
}

func (h *apiHandlers) removeBookFromShelf(w http.ResponseWriter, r *http.Request) {
	segs := extractPathSegments(r.URL.Path)
	// /api/shelves/{id}/books/{bid}
	var shelfID, bookID int64
	for i, s := range segs {
		if s == "shelves" && i+1 < len(segs) {
			shelfID, _ = strconv.ParseInt(segs[i+1], 10, 64)
		}
		if s == "books" && i+1 < len(segs) {
			bookID, _ = strconv.ParseInt(segs[i+1], 10, 64)
		}
	}
	if shelfID == 0 || bookID == 0 {
		writeError(w, http.StatusBadRequest, "invalid shelf or book ID")
		return
	}

	if err := h.lib.RemoveBookFromShelf(shelfID, bookID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// --- Tag handlers ---

func (h *apiHandlers) listTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.lib.ListTags()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tags)
}

func (h *apiHandlers) tagBook(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req models.TagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := h.lib.TagBook(id, req.Name); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "tagged"})
}

func (h *apiHandlers) untagBook(w http.ResponseWriter, r *http.Request) {
	segs := extractPathSegments(r.URL.Path)
	// /api/books/{id}/tags/{tid}
	var bookID, tagID int64
	for i, s := range segs {
		if s == "books" && i+1 < len(segs) {
			bookID, _ = strconv.ParseInt(segs[i+1], 10, 64)
		}
		if s == "tags" && i+1 < len(segs) {
			tagID, _ = strconv.ParseInt(segs[i+1], 10, 64)
		}
	}
	if bookID == 0 || tagID == 0 {
		writeError(w, http.StatusBadRequest, "invalid book or tag ID")
		return
	}

	if err := h.lib.UntagBook(bookID, tagID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "untagged"})
}

// --- Stats handler ---

func (h *apiHandlers) getStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.lib.GetStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// --- Import/Export handlers ---

func (h *apiHandlers) importData(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")

	var result *models.ImportResult
	var err error

	if strings.Contains(ct, "json") {
		result, err = h.lib.ImportJSON(r.Body)
	} else {
		// Default to CSV
		result, err = h.lib.ImportCSV(r.Body)
	}

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *apiHandlers) exportData(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=library.csv")
		if err := h.lib.ExportCSV(w); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=library.json")
		if err := h.lib.ExportJSON(w); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
	}
}

// --- Auth handlers ---

func (h *apiHandlers) login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	resp, err := h.lib.Login(req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    resp.Token,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, resp)
}

func (h *apiHandlers) register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// First user is always admin; subsequent users require admin auth
	isFirstUser := h.lib.IsRegistrationOpen()
	if !isFirstUser {
		// Check if the requester is an admin
		user := UserFromContext(r)
		if user == nil || !user.IsAdmin {
			writeError(w, http.StatusForbidden, "only admins can create new users")
			return
		}
	}

	user, err := h.lib.Register(&req, isFirstUser)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// If first user, auto-login
	if isFirstUser {
		resp, err := h.lib.Login(req.Username, req.Password)
		if err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    resp.Token,
				Path:     "/",
				MaxAge:   30 * 24 * 60 * 60,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			writeJSON(w, http.StatusCreated, resp)
			return
		}
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *apiHandlers) logout(w http.ResponseWriter, r *http.Request) {
	// Get token from cookie or header
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

	if token != "" {
		h.lib.Logout(token)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (h *apiHandlers) me(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *apiHandlers) changePassword(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := h.lib.ChangePassword(user.ID, req.CurrentPassword, req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "password changed"})
}

func (h *apiHandlers) updateProfile(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	updated, err := h.lib.UpdateProfile(user.ID, req.DisplayName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *apiHandlers) listUsers(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r)
	if user == nil || !user.IsAdmin {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	users, err := h.lib.ListUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *apiHandlers) deleteUser(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r)
	if user == nil || !user.IsAdmin {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if id == user.ID {
		writeError(w, http.StatusBadRequest, "cannot delete your own account")
		return
	}

	if err := h.lib.DeleteUser(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- ISBN Lookup handler ---

func (h *apiHandlers) lookupISBN(w http.ResponseWriter, r *http.Request) {
	segs := extractPathSegments(r.URL.Path)
	// /api/lookup/{isbn}
	isbn := ""
	for i, s := range segs {
		if s == "lookup" && i+1 < len(segs) {
			isbn = segs[i+1]
			break
		}
	}
	if isbn == "" {
		writeError(w, http.StatusBadRequest, "ISBN is required")
		return
	}

	result, err := h.lib.LookupISBN(isbn)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// --- Loan handlers ---

func (h *apiHandlers) listLoans(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	params := models.LoanListParams{
		Status:     q.Get("status"),
		LoanType:   q.Get("loan_type"),
		PersonName: q.Get("person"),
	}
	if bid := q.Get("book_id"); bid != "" {
		params.BookID, _ = strconv.ParseInt(bid, 10, 64)
	}
	if lim := q.Get("limit"); lim != "" {
		params.Limit, _ = strconv.Atoi(lim)
	}
	if off := q.Get("offset"); off != "" {
		params.Offset, _ = strconv.Atoi(off)
	}

	result, err := h.lib.ListLoans(params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *apiHandlers) createLoan(w http.ResponseWriter, r *http.Request) {
	var req models.CreateLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	loan, err := h.lib.CheckOut(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, loan)
}

func (h *apiHandlers) getLoan(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	loan, err := h.lib.GetLoan(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "loan not found")
		return
	}
	writeJSON(w, http.StatusOK, loan)
}

func (h *apiHandlers) checkInLoan(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req models.CheckInRequest
	json.NewDecoder(r.Body).Decode(&req) // optional body

	loan, err := h.lib.CheckIn(id, req.Notes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, loan)
}

func (h *apiHandlers) deleteLoan(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.lib.DeleteLoan(id); err != nil {
		writeError(w, http.StatusNotFound, "loan not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *apiHandlers) getBookLoans(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	loans, err := h.lib.GetBookLoanHistory(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, loans)
}
