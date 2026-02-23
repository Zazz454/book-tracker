package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/user/library/internal/models"
)

// --- Books ---

// CreateBook inserts a new book and returns its ID.
func (d *DB) CreateBook(b *models.CreateBookRequest) (int64, error) {
	status := b.Status
	if status == "" {
		status = models.StatusUnread
	}
	res, err := d.Exec(
		`INSERT INTO books (title, author, isbn, genre, pages, year, status, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		b.Title, b.Author, b.ISBN, b.Genre, b.Pages, b.Year, status, b.Notes,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetBook retrieves a single book by ID, including its shelves and tags.
func (d *DB) GetBook(id int64) (*models.Book, error) {
	b := &models.Book{}
	var startedAt, finishedAt sql.NullTime
	err := d.QueryRow(
		`SELECT id, title, author, isbn, genre, pages, year,
		        cover_path, cover_source, status, rating, notes,
		        started_at, finished_at, created_at, updated_at
		 FROM books WHERE id = ?`, id,
	).Scan(
		&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Genre, &b.Pages, &b.Year,
		&b.CoverPath, &b.CoverSource, &b.Status, &b.Rating, &b.Notes,
		&startedAt, &finishedAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		b.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		b.FinishedAt = &finishedAt.Time
	}

	// Load shelves
	b.Shelves, _ = d.GetBookShelves(id)
	// Load tags
	b.Tags, _ = d.GetBookTags(id)

	return b, nil
}

// ListBooks retrieves books with optional filtering, sorting, and pagination.
func (d *DB) ListBooks(p models.BookListParams) (*models.BookListResponse, error) {
	var (
		where  []string
		args   []interface{}
		joins  []string
	)

	if p.Status != "" {
		where = append(where, "b.status = ?")
		args = append(args, p.Status)
	}
	if p.Genre != "" {
		where = append(where, "b.genre = ?")
		args = append(args, p.Genre)
	}
	if p.ShelfID > 0 {
		joins = append(joins, "JOIN book_shelves bs ON bs.book_id = b.id")
		where = append(where, "bs.shelf_id = ?")
		args = append(args, p.ShelfID)
	}
	if p.TagName != "" {
		joins = append(joins, "JOIN book_tags bt ON bt.book_id = b.id JOIN tags t ON t.id = bt.tag_id")
		where = append(where, "t.name = ?")
		args = append(args, p.TagName)
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}
	joinClause := strings.Join(joins, " ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT b.id) FROM books b %s %s", joinClause, whereClause)
	var total int
	if err := d.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Sort
	sortCol := "b.created_at"
	switch p.Sort {
	case "title":
		sortCol = "b.title"
	case "author":
		sortCol = "b.author"
	case "rating":
		sortCol = "b.rating"
	case "year":
		sortCol = "b.year"
	case "updated_at":
		sortCol = "b.updated_at"
	case "pages":
		sortCol = "b.pages"
	case "status":
		sortCol = "b.status"
	}
	order := "DESC"
	if p.Order == "asc" {
		order = "ASC"
	}

	limit := p.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(
		`SELECT DISTINCT b.id, b.title, b.author, b.isbn, b.genre, b.pages, b.year,
		        b.cover_path, b.cover_source, b.status, b.rating, b.notes,
		        b.started_at, b.finished_at, b.created_at, b.updated_at
		 FROM books b %s %s
		 ORDER BY %s %s
		 LIMIT ? OFFSET ?`,
		joinClause, whereClause, sortCol, order,
	)
	args = append(args, limit, offset)

	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		var startedAt, finishedAt sql.NullTime
		if err := rows.Scan(
			&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Genre, &b.Pages, &b.Year,
			&b.CoverPath, &b.CoverSource, &b.Status, &b.Rating, &b.Notes,
			&startedAt, &finishedAt, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			b.StartedAt = &startedAt.Time
		}
		if finishedAt.Valid {
			b.FinishedAt = &finishedAt.Time
		}
		books = append(books, b)
	}
	if books == nil {
		books = []models.Book{}
	}

	return &models.BookListResponse{
		Books:  books,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// UpdateBook updates specific fields of a book.
func (d *DB) UpdateBook(id int64, req *models.UpdateBookRequest) error {
	var sets []string
	var args []interface{}

	if req.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *req.Title)
	}
	if req.Author != nil {
		sets = append(sets, "author = ?")
		args = append(args, *req.Author)
	}
	if req.ISBN != nil {
		sets = append(sets, "isbn = ?")
		args = append(args, *req.ISBN)
	}
	if req.Genre != nil {
		sets = append(sets, "genre = ?")
		args = append(args, *req.Genre)
	}
	if req.Pages != nil {
		sets = append(sets, "pages = ?")
		args = append(args, *req.Pages)
	}
	if req.Year != nil {
		sets = append(sets, "year = ?")
		args = append(args, *req.Year)
	}
	if req.Notes != nil {
		sets = append(sets, "notes = ?")
		args = append(args, *req.Notes)
	}

	if len(sets) == 0 {
		return nil
	}

	sets = append(sets, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE books SET %s WHERE id = ?", strings.Join(sets, ", "))
	res, err := d.Exec(query, args...)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteBook removes a book by ID.
func (d *DB) DeleteBook(id int64) error {
	res, err := d.Exec("DELETE FROM books WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateBookStatus changes a book's reading status and manages timestamps.
func (d *DB) UpdateBookStatus(id int64, status string) error {
	now := time.Now()
	var startedAt, finishedAt interface{}

	switch status {
	case models.StatusReading:
		startedAt = now
	case models.StatusFinished:
		finishedAt = now
	}

	// If setting to reading, only set started_at if not already set
	if status == models.StatusReading {
		var existing sql.NullTime
		d.QueryRow("SELECT started_at FROM books WHERE id = ?", id).Scan(&existing)
		if existing.Valid {
			startedAt = existing.Time
		}
	}

	query := `UPDATE books SET status = ?, updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{status}

	if startedAt != nil {
		query += `, started_at = ?`
		args = append(args, startedAt)
	}
	if finishedAt != nil {
		query += `, finished_at = ?`
		args = append(args, finishedAt)
	}

	query += ` WHERE id = ?`
	args = append(args, id)

	res, err := d.Exec(query, args...)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateBookRating sets a book's rating (0-5).
func (d *DB) UpdateBookRating(id int64, rating int) error {
	res, err := d.Exec(
		`UPDATE books SET rating = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		rating, id,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateBookCover sets a book's cover path and source.
func (d *DB) UpdateBookCover(id int64, coverPath, coverSource string) error {
	_, err := d.Exec(
		`UPDATE books SET cover_path = ?, cover_source = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		coverPath, coverSource, id,
	)
	return err
}

// SearchBooks performs full-text search using FTS5.
func (d *DB) SearchBooks(query string, limit, offset int) (*models.BookListResponse, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Count total matches
	var total int
	err := d.QueryRow(
		`SELECT COUNT(*) FROM books_fts WHERE books_fts MATCH ?`, query,
	).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := d.Query(
		`SELECT b.id, b.title, b.author, b.isbn, b.genre, b.pages, b.year,
		        b.cover_path, b.cover_source, b.status, b.rating, b.notes,
		        b.started_at, b.finished_at, b.created_at, b.updated_at
		 FROM books_fts f
		 JOIN books b ON b.id = f.rowid
		 WHERE books_fts MATCH ?
		 ORDER BY rank
		 LIMIT ? OFFSET ?`,
		query, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		var startedAt, finishedAt sql.NullTime
		if err := rows.Scan(
			&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Genre, &b.Pages, &b.Year,
			&b.CoverPath, &b.CoverSource, &b.Status, &b.Rating, &b.Notes,
			&startedAt, &finishedAt, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			b.StartedAt = &startedAt.Time
		}
		if finishedAt.Valid {
			b.FinishedAt = &finishedAt.Time
		}
		books = append(books, b)
	}
	if books == nil {
		books = []models.Book{}
	}

	return &models.BookListResponse{
		Books:  books,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// --- Shelves ---

// CreateShelf creates a new shelf.
func (d *DB) CreateShelf(req *models.CreateShelfRequest) (int64, error) {
	res, err := d.Exec(
		`INSERT INTO shelves (name, description) VALUES (?, ?)`,
		req.Name, req.Description,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetShelf retrieves a shelf by ID.
func (d *DB) GetShelf(id int64) (*models.Shelf, error) {
	s := &models.Shelf{}
	err := d.QueryRow(
		`SELECT s.id, s.name, s.description, s.created_at,
		        (SELECT COUNT(*) FROM book_shelves WHERE shelf_id = s.id) as book_count
		 FROM shelves s WHERE s.id = ?`, id,
	).Scan(&s.ID, &s.Name, &s.Description, &s.CreatedAt, &s.BookCount)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// ListShelves returns all shelves with book counts.
func (d *DB) ListShelves() ([]models.Shelf, error) {
	rows, err := d.Query(
		`SELECT s.id, s.name, s.description, s.created_at,
		        (SELECT COUNT(*) FROM book_shelves WHERE shelf_id = s.id) as book_count
		 FROM shelves s ORDER BY s.name ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shelves []models.Shelf
	for rows.Next() {
		var s models.Shelf
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.CreatedAt, &s.BookCount); err != nil {
			return nil, err
		}
		shelves = append(shelves, s)
	}
	if shelves == nil {
		shelves = []models.Shelf{}
	}
	return shelves, nil
}

// UpdateShelf updates a shelf's name and/or description.
func (d *DB) UpdateShelf(id int64, req *models.UpdateShelfRequest) error {
	var sets []string
	var args []interface{}
	if req.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *req.Description)
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE shelves SET %s WHERE id = ?", strings.Join(sets, ", "))
	res, err := d.Exec(query, args...)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteShelf removes a shelf by ID.
func (d *DB) DeleteShelf(id int64) error {
	res, err := d.Exec("DELETE FROM shelves WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// AddBookToShelf adds a book to a shelf.
func (d *DB) AddBookToShelf(bookID, shelfID int64) error {
	_, err := d.Exec(
		`INSERT OR IGNORE INTO book_shelves (book_id, shelf_id) VALUES (?, ?)`,
		bookID, shelfID,
	)
	return err
}

// RemoveBookFromShelf removes a book from a shelf.
func (d *DB) RemoveBookFromShelf(bookID, shelfID int64) error {
	_, err := d.Exec(
		`DELETE FROM book_shelves WHERE book_id = ? AND shelf_id = ?`,
		bookID, shelfID,
	)
	return err
}

// GetBookShelves returns all shelves a book belongs to.
func (d *DB) GetBookShelves(bookID int64) ([]models.Shelf, error) {
	rows, err := d.Query(
		`SELECT s.id, s.name, s.description, s.created_at
		 FROM shelves s
		 JOIN book_shelves bs ON bs.shelf_id = s.id
		 WHERE bs.book_id = ?
		 ORDER BY s.name`, bookID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shelves []models.Shelf
	for rows.Next() {
		var s models.Shelf
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.CreatedAt); err != nil {
			return nil, err
		}
		shelves = append(shelves, s)
	}
	if shelves == nil {
		shelves = []models.Shelf{}
	}
	return shelves, nil
}

// --- Tags ---

// CreateTag creates a tag or returns existing one.
func (d *DB) CreateTag(name string) (int64, error) {
	// Try to insert, ignore if exists
	d.Exec("INSERT OR IGNORE INTO tags (name) VALUES (?)", name)

	var id int64
	err := d.QueryRow("SELECT id FROM tags WHERE name = ?", name).Scan(&id)
	return id, err
}

// ListTags returns all tags.
func (d *DB) ListTags() ([]models.Tag, error) {
	rows, err := d.Query("SELECT id, name FROM tags ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	if tags == nil {
		tags = []models.Tag{}
	}
	return tags, nil
}

// TagBook associates a tag with a book.
func (d *DB) TagBook(bookID, tagID int64) error {
	_, err := d.Exec(
		`INSERT OR IGNORE INTO book_tags (book_id, tag_id) VALUES (?, ?)`,
		bookID, tagID,
	)
	return err
}

// UntagBook removes a tag from a book.
func (d *DB) UntagBook(bookID, tagID int64) error {
	_, err := d.Exec(
		`DELETE FROM book_tags WHERE book_id = ? AND tag_id = ?`,
		bookID, tagID,
	)
	return err
}

// GetBookTags returns all tags for a book.
func (d *DB) GetBookTags(bookID int64) ([]models.Tag, error) {
	rows, err := d.Query(
		`SELECT t.id, t.name FROM tags t
		 JOIN book_tags bt ON bt.tag_id = t.id
		 WHERE bt.book_id = ?
		 ORDER BY t.name`, bookID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	if tags == nil {
		tags = []models.Tag{}
	}
	return tags, nil
}

// --- Statistics ---

// GetStats returns reading statistics.
func (d *DB) GetStats() (*models.StatsResponse, error) {
	s := &models.StatsResponse{}

	// Totals by status
	d.QueryRow("SELECT COUNT(*) FROM books").Scan(&s.TotalBooks)
	d.QueryRow("SELECT COUNT(*) FROM books WHERE status = ?", models.StatusFinished).Scan(&s.BooksRead)
	d.QueryRow("SELECT COUNT(*) FROM books WHERE status = ?", models.StatusReading).Scan(&s.BooksReading)
	d.QueryRow("SELECT COUNT(*) FROM books WHERE status = ?", models.StatusUnread).Scan(&s.BooksUnread)

	// Total pages (finished books)
	d.QueryRow("SELECT COALESCE(SUM(pages), 0) FROM books WHERE status = ?", models.StatusFinished).Scan(&s.TotalPages)

	// Average rating (of rated books)
	d.QueryRow("SELECT COALESCE(AVG(CAST(rating AS REAL)), 0) FROM books WHERE rating > 0").Scan(&s.AverageRating)

	// Average days per book
	d.QueryRow(`SELECT COALESCE(AVG(julianday(finished_at) - julianday(started_at)), 0)
		FROM books WHERE status = 'finished' AND started_at IS NOT NULL AND finished_at IS NOT NULL`).Scan(&s.AvgDaysPerBook)

	// Books per month (last 12 months)
	rows, err := d.Query(`
		SELECT strftime('%Y-%m', finished_at) as month, COUNT(*) as count
		FROM books
		WHERE status = 'finished' AND finished_at IS NOT NULL
		GROUP BY month
		ORDER BY month DESC
		LIMIT 12`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var mc models.MonthCount
			rows.Scan(&mc.Month, &mc.Count)
			s.BooksPerMonth = append(s.BooksPerMonth, mc)
		}
	}
	if s.BooksPerMonth == nil {
		s.BooksPerMonth = []models.MonthCount{}
	}

	// Books per year
	rows2, err := d.Query(`
		SELECT CAST(strftime('%Y', finished_at) AS INTEGER) as year, COUNT(*) as count
		FROM books
		WHERE status = 'finished' AND finished_at IS NOT NULL
		GROUP BY year
		ORDER BY year DESC`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var yc models.YearCount
			rows2.Scan(&yc.Year, &yc.Count)
			s.BooksPerYear = append(s.BooksPerYear, yc)
		}
	}
	if s.BooksPerYear == nil {
		s.BooksPerYear = []models.YearCount{}
	}

	// Genre breakdown
	rows3, err := d.Query(`
		SELECT genre, COUNT(*) as count
		FROM books
		WHERE genre != ''
		GROUP BY genre
		ORDER BY count DESC`)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var gc models.GenreCount
			rows3.Scan(&gc.Genre, &gc.Count)
			s.GenreBreakdown = append(s.GenreBreakdown, gc)
		}
	}
	if s.GenreBreakdown == nil {
		s.GenreBreakdown = []models.GenreCount{}
	}

	// Rating distribution
	rows4, err := d.Query(`
		SELECT rating, COUNT(*) as count
		FROM books
		WHERE rating > 0
		GROUP BY rating
		ORDER BY rating DESC`)
	if err == nil {
		defer rows4.Close()
		for rows4.Next() {
			var rc models.RatingCount
			rows4.Scan(&rc.Rating, &rc.Count)
			s.RatingDist = append(s.RatingDist, rc)
		}
	}
	if s.RatingDist == nil {
		s.RatingDist = []models.RatingCount{}
	}

	// Longest and shortest finished books
	longestBook := &models.Book{}
	err = d.QueryRow(`SELECT id, title, author, pages FROM books
		WHERE status = 'finished' AND pages > 0
		ORDER BY pages DESC LIMIT 1`).Scan(&longestBook.ID, &longestBook.Title, &longestBook.Author, &longestBook.Pages)
	if err == nil {
		s.LongestBook = longestBook
	}

	shortestBook := &models.Book{}
	err = d.QueryRow(`SELECT id, title, author, pages FROM books
		WHERE status = 'finished' AND pages > 0
		ORDER BY pages ASC LIMIT 1`).Scan(&shortestBook.ID, &shortestBook.Title, &shortestBook.Author, &shortestBook.Pages)
	if err == nil {
		s.ShortestBook = shortestBook
	}

	return s, nil
}

// GetAllBooks returns all books (used for export).
func (d *DB) GetAllBooks() ([]models.Book, error) {
	rows, err := d.Query(
		`SELECT id, title, author, isbn, genre, pages, year,
		        cover_path, cover_source, status, rating, notes,
		        started_at, finished_at, created_at, updated_at
		 FROM books ORDER BY id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		var startedAt, finishedAt sql.NullTime
		if err := rows.Scan(
			&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Genre, &b.Pages, &b.Year,
			&b.CoverPath, &b.CoverSource, &b.Status, &b.Rating, &b.Notes,
			&startedAt, &finishedAt, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			b.StartedAt = &startedAt.Time
		}
		if finishedAt.Valid {
			b.FinishedAt = &finishedAt.Time
		}
		books = append(books, b)
	}
	if books == nil {
		books = []models.Book{}
	}
	return books, nil
}
