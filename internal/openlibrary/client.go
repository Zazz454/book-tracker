package openlibrary

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/user/library/internal/models"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// booksAPIResponse is the raw response from the Open Library Books API.
type booksAPIResponse map[string]struct {
	Title      string `json:"title"`
	Authors    []struct {
		Name string `json:"name"`
	} `json:"authors"`
	Publishers []struct {
		Name string `json:"name"`
	} `json:"publishers"`
	NumberOfPages int    `json:"number_of_pages"`
	PublishDate   string `json:"publish_date"`
	Subjects      []struct {
		Name string `json:"name"`
	} `json:"subjects"`
	Cover struct {
		Small  string `json:"small"`
		Medium string `json:"medium"`
		Large  string `json:"large"`
	} `json:"cover"`
}

// LookupISBN fetches book metadata from Open Library by ISBN.
func LookupISBN(isbn string) (*models.ISBNLookupResult, error) {
	isbn = strings.TrimSpace(isbn)
	isbn = strings.ReplaceAll(isbn, "-", "")

	apiURL := fmt.Sprintf("https://openlibrary.org/api/books?bibkeys=ISBN:%s&format=json&jscmd=data", url.QueryEscape(isbn))

	resp, err := httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("openlibrary lookup: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openlibrary lookup: status %d", resp.StatusCode)
	}

	var raw booksAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("openlibrary decode: %w", err)
	}

	key := "ISBN:" + isbn
	data, ok := raw[key]
	if !ok {
		return nil, fmt.Errorf("openlibrary: no data found for ISBN %s", isbn)
	}

	result := &models.ISBNLookupResult{
		Title: data.Title,
		ISBN:  isbn,
		Pages: data.NumberOfPages,
	}

	// Authors
	var authors []string
	for _, a := range data.Authors {
		authors = append(authors, a.Name)
	}
	result.Author = strings.Join(authors, ", ")

	// Publisher
	if len(data.Publishers) > 0 {
		result.Publisher = data.Publishers[0].Name
	}

	// Year from publish date
	if data.PublishDate != "" {
		result.Year = parseYear(data.PublishDate)
	}

	// Subjects -> genre (take first one)
	for _, s := range data.Subjects {
		result.Subjects = append(result.Subjects, s.Name)
	}
	if len(data.Subjects) > 0 {
		result.Genre = data.Subjects[0].Name
	}

	// Cover URL
	if data.Cover.Large != "" {
		result.CoverURL = data.Cover.Large
	} else if data.Cover.Medium != "" {
		result.CoverURL = data.Cover.Medium
	} else if data.Cover.Small != "" {
		result.CoverURL = data.Cover.Small
	}

	return result, nil
}

// searchResponse is the Open Library search response structure.
type searchResponse struct {
	Docs []struct {
		Title         string   `json:"title"`
		AuthorName    []string `json:"author_name"`
		ISBN          []string `json:"isbn"`
		FirstPublish  int      `json:"first_publish_year"`
		NumberOfPages int      `json:"number_of_pages_median"`
		Subject       []string `json:"subject"`
		CoverI        int      `json:"cover_i"`
	} `json:"docs"`
}

// SearchByTitle searches Open Library by title and author.
func SearchByTitle(title, author string) (*models.ISBNLookupResult, error) {
	q := url.Values{}
	q.Set("title", title)
	if author != "" {
		q.Set("author", author)
	}
	q.Set("fields", "title,author_name,isbn,first_publish_year,number_of_pages_median,subject,cover_i")
	q.Set("limit", "1")

	searchURL := "https://openlibrary.org/search.json?" + q.Encode()
	resp, err := httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("openlibrary search: %w", err)
	}
	defer resp.Body.Close()

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}

	if len(sr.Docs) == 0 {
		return nil, fmt.Errorf("openlibrary: no results for %q", title)
	}

	doc := sr.Docs[0]
	result := &models.ISBNLookupResult{
		Title: doc.Title,
		Year:  doc.FirstPublish,
		Pages: doc.NumberOfPages,
	}

	if len(doc.AuthorName) > 0 {
		result.Author = strings.Join(doc.AuthorName, ", ")
	}
	if len(doc.ISBN) > 0 {
		result.ISBN = doc.ISBN[0]
	}
	if len(doc.Subject) > 0 {
		result.Genre = doc.Subject[0]
		for _, s := range doc.Subject {
			result.Subjects = append(result.Subjects, s)
		}
	}
	if doc.CoverI > 0 {
		result.CoverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", doc.CoverI)
	}

	return result, nil
}

// parseYear attempts to extract a 4-digit year from a date string.
func parseYear(s string) int {
	s = strings.TrimSpace(s)
	// Try last 4 chars, or find a 4-digit number
	for _, part := range strings.Fields(s) {
		part = strings.TrimRight(part, ",.")
		if len(part) == 4 {
			if y, err := strconv.Atoi(part); err == nil && y > 1000 && y < 2100 {
				return y
			}
		}
	}
	return 0
}
