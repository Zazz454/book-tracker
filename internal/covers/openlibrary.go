package covers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// FetchOpenLibraryByISBN tries to download a cover image by ISBN from Open Library.
// Returns the image data and content type, or an error if not found.
func FetchOpenLibraryByISBN(isbn string) ([]byte, string, error) {
	coverURL := fmt.Sprintf("https://covers.openlibrary.org/b/isbn/%s-L.jpg?default=false", isbn)

	resp, err := httpClient.Get(coverURL)
	if err != nil {
		return nil, "", fmt.Errorf("openlibrary cover request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("openlibrary cover: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10MB max
	if err != nil {
		return nil, "", fmt.Errorf("openlibrary cover read: %w", err)
	}

	// Open Library returns a 1x1 pixel for missing covers sometimes
	if len(data) < 1000 {
		return nil, "", fmt.Errorf("openlibrary cover: image too small, likely placeholder")
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}
	return data, ct, nil
}

// openLibrarySearchResult is the structure returned by the OL search API.
type openLibrarySearchResult struct {
	Docs []struct {
		CoverI    int      `json:"cover_i"`
		Title     string   `json:"title"`
		AuthorName []string `json:"author_name"`
	} `json:"docs"`
}

// FetchOpenLibraryBySearch tries to find a cover by searching title+author.
func FetchOpenLibraryBySearch(title, author string) ([]byte, string, error) {
	q := url.Values{}
	q.Set("title", title)
	if author != "" {
		q.Set("author", author)
	}
	q.Set("fields", "cover_i,title,author_name")
	q.Set("limit", "1")

	searchURL := "https://openlibrary.org/search.json?" + q.Encode()
	resp, err := httpClient.Get(searchURL)
	if err != nil {
		return nil, "", fmt.Errorf("openlibrary search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("openlibrary search: status %d", resp.StatusCode)
	}

	var result openLibrarySearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("openlibrary search decode: %w", err)
	}

	if len(result.Docs) == 0 || result.Docs[0].CoverI == 0 {
		return nil, "", fmt.Errorf("openlibrary search: no cover found for %q by %q", title, author)
	}

	coverURL := fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", result.Docs[0].CoverI)
	coverResp, err := httpClient.Get(coverURL)
	if err != nil {
		return nil, "", fmt.Errorf("openlibrary cover download: %w", err)
	}
	defer coverResp.Body.Close()

	if coverResp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("openlibrary cover download: status %d", coverResp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(coverResp.Body, 10*1024*1024))
	if err != nil {
		return nil, "", err
	}

	if len(data) < 1000 {
		return nil, "", fmt.Errorf("openlibrary: image too small")
	}

	_ = strings.TrimSpace // keep import
	return data, "image/jpeg", nil
}
