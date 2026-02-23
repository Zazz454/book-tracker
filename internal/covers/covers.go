package covers

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"
)

const (
	maxWidth  = 300
	maxHeight = 450
)

// Fetcher handles downloading and saving cover images.
type Fetcher struct {
	coversDir string
}

// NewFetcher creates a new cover fetcher that saves images to the given directory.
func NewFetcher(coversDir string) *Fetcher {
	return &Fetcher{coversDir: coversDir}
}

// FetchAndSave attempts to download a cover image for a book and save it locally.
// It tries sources in order: Open Library by ISBN, Open Library by search, Google Books.
// Returns the relative cover path and source name, or an error if all sources fail.
func (f *Fetcher) FetchAndSave(bookID int64, isbn, title, author string) (string, string, error) {
	var imgData []byte
	var contentType string
	var source string
	var err error

	// 1. Try Open Library by ISBN
	if isbn != "" {
		imgData, contentType, err = FetchOpenLibraryByISBN(isbn)
		if err == nil {
			source = "openlibrary"
		}
	}

	// 2. Try Open Library by search
	if imgData == nil {
		imgData, contentType, err = FetchOpenLibraryBySearch(title, author)
		if err == nil {
			source = "openlibrary"
		}
	}

	// 3. Try Google Books
	if imgData == nil {
		imgData, contentType, err = FetchGoogleBooks(title, author)
		if err == nil {
			source = "google"
		}
	}

	if imgData == nil {
		return "", "", fmt.Errorf("no cover found from any source")
	}

	// Process and save the image
	processed, procErr := processImage(imgData, contentType)
	if procErr != nil {
		// If processing fails, save the raw image
		log.Printf("cover processing failed for book %d, saving raw: %v", bookID, procErr)
		processed = imgData
	}

	filename := fmt.Sprintf("%d.jpg", bookID)
	coverPath := filepath.Join(f.coversDir, filename)

	if err := os.WriteFile(coverPath, processed, 0644); err != nil {
		return "", "", fmt.Errorf("save cover: %w", err)
	}

	return filepath.Join("covers", filename), source, nil
}

// DownloadURL downloads an image from a URL and saves it for the given book ID.
func (f *Fetcher) DownloadURL(bookID int64, imageURL string) (string, error) {
	resp, err := httpClient.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("download cover: %w", err)
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return "", fmt.Errorf("read cover: %w", err)
	}

	ct := resp.Header.Get("Content-Type")
	processed, procErr := processImage(buf.Bytes(), ct)
	if procErr != nil {
		processed = buf.Bytes()
	}

	filename := fmt.Sprintf("%d.jpg", bookID)
	coverPath := filepath.Join(f.coversDir, filename)

	if err := os.WriteFile(coverPath, processed, 0644); err != nil {
		return "", fmt.Errorf("save cover: %w", err)
	}

	return filepath.Join("covers", filename), nil
}

// processImage decodes, resizes (if needed), and re-encodes an image as JPEG.
func processImage(data []byte, contentType string) ([]byte, error) {
	var img image.Image
	var err error

	reader := bytes.NewReader(data)

	switch {
	case contentType == "image/png" || bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}):
		img, err = png.Decode(reader)
	default:
		img, err = jpeg.Decode(reader)
	}

	if err != nil {
		// Try generic decode as fallback
		reader.Reset(data)
		img, _, err = image.Decode(reader)
		if err != nil {
			return nil, fmt.Errorf("decode image: %w", err)
		}
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	// Resize if larger than max dimensions
	if w > maxWidth || h > maxHeight {
		ratio := float64(w) / float64(h)
		newW := maxWidth
		newH := int(float64(newW) / ratio)
		if newH > maxHeight {
			newH = maxHeight
			newW = int(float64(newH) * ratio)
		}

		dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
		draw.BiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
		img = dst
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("encode jpeg: %w", err)
	}

	return buf.Bytes(), nil
}
