package covers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// googleBooksResponse is the structure returned by the Google Books API.
type googleBooksResponse struct {
	TotalItems int `json:"totalItems"`
	Items      []struct {
		VolumeInfo struct {
			Title      string `json:"title"`
			ImageLinks struct {
				Thumbnail      string `json:"thumbnail"`
				SmallThumbnail string `json:"smallThumbnail"`
			} `json:"imageLinks"`
		} `json:"volumeInfo"`
	} `json:"items"`
}

// FetchGoogleBooks tries to find and download a cover from Google Books API.
func FetchGoogleBooks(title, author string) ([]byte, string, error) {
	q := title
	if author != "" {
		q += "+inauthor:" + author
	}

	searchURL := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=%s&maxResults=1&fields=totalItems,items/volumeInfo/title,items/volumeInfo/imageLinks",
		url.QueryEscape(q))

	resp, err := httpClient.Get(searchURL)
	if err != nil {
		return nil, "", fmt.Errorf("google books search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("google books search: status %d", resp.StatusCode)
	}

	var result googleBooksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("google books decode: %w", err)
	}

	if result.TotalItems == 0 || len(result.Items) == 0 {
		return nil, "", fmt.Errorf("google books: no results for %q by %q", title, author)
	}

	imageURL := result.Items[0].VolumeInfo.ImageLinks.Thumbnail
	if imageURL == "" {
		imageURL = result.Items[0].VolumeInfo.ImageLinks.SmallThumbnail
	}
	if imageURL == "" {
		return nil, "", fmt.Errorf("google books: no cover image available")
	}

	// Google Books returns HTTP URLs; upgrade to HTTPS
	imageURL = strings.Replace(imageURL, "http://", "https://", 1)

	// Download the image
	imgResp, err := httpClient.Get(imageURL)
	if err != nil {
		return nil, "", fmt.Errorf("google books cover download: %w", err)
	}
	defer imgResp.Body.Close()

	if imgResp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("google books cover: status %d", imgResp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(imgResp.Body, 10*1024*1024))
	if err != nil {
		return nil, "", err
	}

	if len(data) < 500 {
		return nil, "", fmt.Errorf("google books: image too small")
	}

	ct := imgResp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}
	return data, ct, nil
}
