package models

import (
	"time"
)

// Book reading status constants
const (
	StatusUnread    = "unread"
	StatusReading   = "reading"
	StatusFinished  = "finished"
	StatusAbandoned = "abandoned"
)

// Cover source constants
const (
	CoverSourceOpenLibrary = "openlibrary"
	CoverSourceGoogle      = "google"
	CoverSourceManual      = "manual"
)

// Book represents a book in the library.
type Book struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Author      string     `json:"author"`
	ISBN        string     `json:"isbn,omitempty"`
	Genre       string     `json:"genre,omitempty"`
	Pages       int        `json:"pages,omitempty"`
	Year        int        `json:"year,omitempty"`
	CoverPath   string     `json:"cover_path,omitempty"`
	CoverSource string     `json:"cover_source,omitempty"`
	Status      string     `json:"status"`
	Rating      int        `json:"rating"`
	Notes       string     `json:"notes,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	// Joined data (not always populated)
	Shelves []Shelf `json:"shelves,omitempty"`
	Tags    []Tag   `json:"tags,omitempty"`
}

// Shelf represents a user-defined collection of books.
type Shelf struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	BookCount   int       `json:"book_count,omitempty"`
}

// Tag represents a label that can be applied to books.
type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// --- Request types ---

// CreateBookRequest is the payload for creating a new book.
type CreateBookRequest struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	ISBN   string `json:"isbn,omitempty"`
	Genre  string `json:"genre,omitempty"`
	Pages  int    `json:"pages,omitempty"`
	Year   int    `json:"year,omitempty"`
	Status string `json:"status,omitempty"`
	Notes  string `json:"notes,omitempty"`
}

// UpdateBookRequest is the payload for updating an existing book.
type UpdateBookRequest struct {
	Title  *string `json:"title,omitempty"`
	Author *string `json:"author,omitempty"`
	ISBN   *string `json:"isbn,omitempty"`
	Genre  *string `json:"genre,omitempty"`
	Pages  *int    `json:"pages,omitempty"`
	Year   *int    `json:"year,omitempty"`
	Notes  *string `json:"notes,omitempty"`
}

// UpdateStatusRequest is the payload for changing reading status.
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// UpdateRatingRequest is the payload for rating a book.
type UpdateRatingRequest struct {
	Rating int `json:"rating"`
}

// CreateShelfRequest is the payload for creating a shelf.
type CreateShelfRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateShelfRequest is the payload for updating a shelf.
type UpdateShelfRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ShelfBookRequest is the payload for adding a book to a shelf.
type ShelfBookRequest struct {
	BookID int64 `json:"book_id"`
}

// TagRequest is the payload for tagging a book.
type TagRequest struct {
	Name string `json:"name"`
}

// SetCoverRequest is the payload for setting a custom cover URL.
type SetCoverRequest struct {
	URL string `json:"url"`
}

// --- Query/filter types ---

// BookListParams controls filtering, sorting, and pagination of book lists.
type BookListParams struct {
	Status  string `json:"status,omitempty"`
	Genre   string `json:"genre,omitempty"`
	ShelfID int64  `json:"shelf_id,omitempty"`
	TagName string `json:"tag,omitempty"`
	Sort    string `json:"sort,omitempty"`  // title, author, rating, year, created_at, updated_at
	Order   string `json:"order,omitempty"` // asc, desc
	Query   string `json:"q,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Offset  int    `json:"offset,omitempty"`
}

// --- Response types ---

// BookListResponse wraps a list of books with pagination info.
type BookListResponse struct {
	Books  []Book `json:"books"`
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// StatsResponse contains reading statistics.
type StatsResponse struct {
	TotalBooks     int              `json:"total_books"`
	BooksRead      int              `json:"books_read"`
	BooksReading   int              `json:"books_reading"`
	BooksUnread    int              `json:"books_unread"`
	TotalPages     int              `json:"total_pages"`
	AverageRating  float64          `json:"average_rating"`
	AvgDaysPerBook float64          `json:"avg_days_per_book"`
	BooksPerMonth  []MonthCount     `json:"books_per_month"`
	BooksPerYear   []YearCount      `json:"books_per_year"`
	GenreBreakdown []GenreCount     `json:"genre_breakdown"`
	RatingDist     []RatingCount    `json:"rating_distribution"`
	ReadingStreak  int              `json:"reading_streak"`
	LongestBook    *Book            `json:"longest_book,omitempty"`
	ShortestBook   *Book            `json:"shortest_book,omitempty"`
}

// MonthCount is a count of books for a given month.
type MonthCount struct {
	Month string `json:"month"` // YYYY-MM
	Count int    `json:"count"`
}

// YearCount is a count of books for a given year.
type YearCount struct {
	Year  int `json:"year"`
	Count int `json:"count"`
}

// GenreCount is a count of books per genre.
type GenreCount struct {
	Genre string `json:"genre"`
	Count int    `json:"count"`
}

// RatingCount is a count of books per rating value.
type RatingCount struct {
	Rating int `json:"rating"`
	Count  int `json:"count"`
}

// ISBNLookupResult is the result of an ISBN metadata lookup.
type ISBNLookupResult struct {
	Title      string   `json:"title"`
	Author     string   `json:"author"`
	ISBN       string   `json:"isbn"`
	Genre      string   `json:"genre,omitempty"`
	Pages      int      `json:"pages,omitempty"`
	Year       int      `json:"year,omitempty"`
	CoverURL   string   `json:"cover_url,omitempty"`
	Publisher  string   `json:"publisher,omitempty"`
	Subjects   []string `json:"subjects,omitempty"`
}

// ImportResult summarizes the result of a bulk import.
type ImportResult struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}
