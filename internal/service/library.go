package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/user/library/internal/covers"
	"github.com/user/library/internal/db"
	"github.com/user/library/internal/models"
	"github.com/user/library/internal/openlibrary"
)

// Library is the core service that coordinates between DB, covers, and external APIs.
type Library struct {
	DB      *db.DB
	Covers  *covers.Fetcher
}

// NewLibrary creates a new Library service.
func NewLibrary(database *db.DB) *Library {
	return &Library{
		DB:     database,
		Covers: covers.NewFetcher(database.CoversDir()),
	}
}

// --- Book operations ---

// CreateBook creates a new book and triggers async cover fetching.
func (l *Library) CreateBook(req *models.CreateBookRequest) (*models.Book, error) {
	if err := validateCreateBook(req); err != nil {
		return nil, err
	}

	id, err := l.DB.CreateBook(req)
	if err != nil {
		return nil, fmt.Errorf("create book: %w", err)
	}

	// Fetch cover asynchronously
	go l.fetchCover(id, req.ISBN, req.Title, req.Author)

	return l.DB.GetBook(id)
}

// GetBook retrieves a book by ID.
func (l *Library) GetBook(id int64) (*models.Book, error) {
	return l.DB.GetBook(id)
}

// ListBooks lists books with filtering and pagination.
func (l *Library) ListBooks(params models.BookListParams) (*models.BookListResponse, error) {
	return l.DB.ListBooks(params)
}

// UpdateBook updates a book's fields.
func (l *Library) UpdateBook(id int64, req *models.UpdateBookRequest) (*models.Book, error) {
	if err := l.DB.UpdateBook(id, req); err != nil {
		return nil, err
	}
	return l.DB.GetBook(id)
}

// DeleteBook deletes a book.
func (l *Library) DeleteBook(id int64) error {
	return l.DB.DeleteBook(id)
}

// UpdateStatus changes a book's reading status.
func (l *Library) UpdateStatus(id int64, status string) (*models.Book, error) {
	if !isValidStatus(status) {
		return nil, fmt.Errorf("invalid status: %q (must be unread, reading, finished, or abandoned)", status)
	}
	if err := l.DB.UpdateBookStatus(id, status); err != nil {
		return nil, err
	}
	return l.DB.GetBook(id)
}

// UpdateRating sets a book's rating.
func (l *Library) UpdateRating(id int64, rating int) (*models.Book, error) {
	if rating < 0 || rating > 5 {
		return nil, fmt.Errorf("invalid rating: %d (must be 0-5)", rating)
	}
	if err := l.DB.UpdateBookRating(id, rating); err != nil {
		return nil, err
	}
	return l.DB.GetBook(id)
}

// SetCoverURL downloads a cover from a URL and sets it for a book.
func (l *Library) SetCoverURL(id int64, imageURL string) (*models.Book, error) {
	coverPath, err := l.Covers.DownloadURL(id, imageURL)
	if err != nil {
		return nil, fmt.Errorf("download cover: %w", err)
	}
	if err := l.DB.UpdateBookCover(id, coverPath, models.CoverSourceManual); err != nil {
		return nil, err
	}
	return l.DB.GetBook(id)
}

// RefreshCover re-fetches a cover from external sources.
func (l *Library) RefreshCover(id int64) (*models.Book, error) {
	book, err := l.DB.GetBook(id)
	if err != nil {
		return nil, err
	}
	go l.fetchCover(id, book.ISBN, book.Title, book.Author)
	return book, nil
}

// SearchBooks performs full-text search.
func (l *Library) SearchBooks(query string, limit, offset int) (*models.BookListResponse, error) {
	if query == "" {
		return &models.BookListResponse{Books: []models.Book{}}, nil
	}
	return l.DB.SearchBooks(query, limit, offset)
}

// --- Shelf operations ---

// CreateShelf creates a new shelf.
func (l *Library) CreateShelf(req *models.CreateShelfRequest) (*models.Shelf, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("shelf name is required")
	}
	id, err := l.DB.CreateShelf(req)
	if err != nil {
		return nil, err
	}
	return l.DB.GetShelf(id)
}

// ListShelves returns all shelves.
func (l *Library) ListShelves() ([]models.Shelf, error) {
	return l.DB.ListShelves()
}

// GetShelf retrieves a shelf by ID.
func (l *Library) GetShelf(id int64) (*models.Shelf, error) {
	return l.DB.GetShelf(id)
}

// UpdateShelf updates a shelf.
func (l *Library) UpdateShelf(id int64, req *models.UpdateShelfRequest) (*models.Shelf, error) {
	if err := l.DB.UpdateShelf(id, req); err != nil {
		return nil, err
	}
	return l.DB.GetShelf(id)
}

// DeleteShelf removes a shelf.
func (l *Library) DeleteShelf(id int64) error {
	return l.DB.DeleteShelf(id)
}

// AddBookToShelf adds a book to a shelf.
func (l *Library) AddBookToShelf(shelfID, bookID int64) error {
	return l.DB.AddBookToShelf(bookID, shelfID)
}

// RemoveBookFromShelf removes a book from a shelf.
func (l *Library) RemoveBookFromShelf(shelfID, bookID int64) error {
	return l.DB.RemoveBookFromShelf(bookID, shelfID)
}

// GetShelfBooks returns books in a shelf.
func (l *Library) GetShelfBooks(shelfID int64) (*models.BookListResponse, error) {
	return l.DB.ListBooks(models.BookListParams{ShelfID: shelfID, Limit: 1000})
}

// --- Tag operations ---

// TagBook adds a tag to a book, creating the tag if necessary.
func (l *Library) TagBook(bookID int64, tagName string) error {
	tagName = strings.TrimSpace(strings.ToLower(tagName))
	if tagName == "" {
		return fmt.Errorf("tag name is required")
	}
	tagID, err := l.DB.CreateTag(tagName)
	if err != nil {
		return err
	}
	return l.DB.TagBook(bookID, tagID)
}

// UntagBook removes a tag from a book.
func (l *Library) UntagBook(bookID, tagID int64) error {
	return l.DB.UntagBook(bookID, tagID)
}

// ListTags returns all tags.
func (l *Library) ListTags() ([]models.Tag, error) {
	return l.DB.ListTags()
}

// --- Statistics ---

// GetStats returns reading statistics.
func (l *Library) GetStats() (*models.StatsResponse, error) {
	return l.DB.GetStats()
}

// --- ISBN Lookup ---

// LookupISBN fetches book metadata from Open Library.
func (l *Library) LookupISBN(isbn string) (*models.ISBNLookupResult, error) {
	return openlibrary.LookupISBN(isbn)
}

// --- Import/Export ---

// ExportJSON exports all books as JSON.
func (l *Library) ExportJSON(w io.Writer) error {
	books, err := l.DB.GetAllBooks()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(books)
}

// ExportCSV exports all books as CSV.
func (l *Library) ExportCSV(w io.Writer) error {
	books, err := l.DB.GetAllBooks()
	if err != nil {
		return err
	}

	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Header
	cw.Write([]string{
		"ID", "Title", "Author", "ISBN", "Genre", "Pages", "Year",
		"Status", "Rating", "Notes", "Started At", "Finished At",
	})

	for _, b := range books {
		started := ""
		if b.StartedAt != nil {
			started = b.StartedAt.Format("2006-01-02")
		}
		finished := ""
		if b.FinishedAt != nil {
			finished = b.FinishedAt.Format("2006-01-02")
		}
		cw.Write([]string{
			strconv.FormatInt(b.ID, 10),
			b.Title, b.Author, b.ISBN, b.Genre,
			strconv.Itoa(b.Pages), strconv.Itoa(b.Year),
			b.Status, strconv.Itoa(b.Rating), b.Notes,
			started, finished,
		})
	}

	return nil
}

// ImportJSON imports books from a JSON array.
func (l *Library) ImportJSON(r io.Reader) (*models.ImportResult, error) {
	var books []models.CreateBookRequest
	if err := json.NewDecoder(r).Decode(&books); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	result := &models.ImportResult{}
	for i, b := range books {
		if b.Title == "" || b.Author == "" {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: missing title or author", i+1))
			continue
		}
		if _, err := l.DB.CreateBook(&b); err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", i+1, err))
			continue
		}
		result.Imported++
	}

	return result, nil
}

// ImportCSV imports books from a CSV file.
func (l *Library) ImportCSV(r io.Reader) (*models.ImportResult, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true
	cr.LazyQuotes = true

	records, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invalid CSV: %w", err)
	}

	if len(records) < 2 {
		return &models.ImportResult{}, nil
	}

	// Build header index map
	header := records[0]
	idx := make(map[string]int)
	for i, h := range header {
		idx[strings.TrimSpace(strings.ToLower(h))] = i
	}

	result := &models.ImportResult{}
	for i, row := range records[1:] {
		b := models.CreateBookRequest{
			Title:  getCSVField(row, idx, "title"),
			Author: getCSVField(row, idx, "author"),
			ISBN:   getCSVField(row, idx, "isbn"),
			Genre:  getCSVField(row, idx, "genre"),
			Status: getCSVField(row, idx, "status"),
			Notes:  getCSVField(row, idx, "notes"),
		}

		// Goodreads compatibility
		if b.Title == "" {
			b.Title = getCSVField(row, idx, "book title")
		}
		if b.Author == "" {
			b.Author = getCSVField(row, idx, "author l-f")
		}
		if b.ISBN == "" {
			isbn := getCSVField(row, idx, "isbn13")
			if isbn == "" {
				isbn = getCSVField(row, idx, "isbn")
			}
			b.ISBN = strings.Trim(isbn, "=\"")
		}
		if b.Status == "" {
			shelf := strings.ToLower(getCSVField(row, idx, "exclusive shelf"))
			switch shelf {
			case "read":
				b.Status = models.StatusFinished
			case "currently-reading":
				b.Status = models.StatusReading
			case "to-read":
				b.Status = models.StatusUnread
			}
		}

		if pages := getCSVField(row, idx, "pages"); pages != "" {
			b.Pages, _ = strconv.Atoi(pages)
		}
		if pages := getCSVField(row, idx, "number of pages"); pages != "" {
			b.Pages, _ = strconv.Atoi(pages)
		}
		if year := getCSVField(row, idx, "year"); year != "" {
			b.Year, _ = strconv.Atoi(year)
		}
		if year := getCSVField(row, idx, "year published"); year != "" {
			b.Year, _ = strconv.Atoi(year)
		}

		if b.Title == "" || b.Author == "" {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: missing title or author", i+2))
			continue
		}

		id, err := l.DB.CreateBook(&b)
		if err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", i+2, err))
			continue
		}

		// Handle rating from CSV
		if rating := getCSVField(row, idx, "rating"); rating != "" {
			if r, err := strconv.Atoi(rating); err == nil && r >= 1 && r <= 5 {
				l.DB.UpdateBookRating(id, r)
			}
		}
		if rating := getCSVField(row, idx, "my rating"); rating != "" {
			if r, err := strconv.Atoi(rating); err == nil && r >= 1 && r <= 5 {
				l.DB.UpdateBookRating(id, r)
			}
		}

		result.Imported++
	}

	return result, nil
}

// --- internal helpers ---

func (l *Library) fetchCover(bookID int64, isbn, title, author string) {
	coverPath, source, err := l.Covers.FetchAndSave(bookID, isbn, title, author)
	if err != nil {
		log.Printf("cover fetch failed for book %d (%s): %v", bookID, title, err)
		return
	}
	if err := l.DB.UpdateBookCover(bookID, coverPath, source); err != nil {
		log.Printf("cover update failed for book %d: %v", bookID, err)
	}
	log.Printf("cover fetched for book %d (%s) from %s", bookID, title, source)
}

func validateCreateBook(req *models.CreateBookRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if strings.TrimSpace(req.Author) == "" {
		return fmt.Errorf("author is required")
	}
	if req.Status != "" && !isValidStatus(req.Status) {
		return fmt.Errorf("invalid status: %q", req.Status)
	}
	return nil
}

func isValidStatus(s string) bool {
	switch s {
	case models.StatusUnread, models.StatusReading, models.StatusFinished, models.StatusAbandoned:
		return true
	}
	return false
}

func getCSVField(row []string, idx map[string]int, field string) string {
	if i, ok := idx[field]; ok && i < len(row) {
		return strings.TrimSpace(row[i])
	}
	return ""
}
