package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"

	"github.com/user/library/internal/service"
)

// Server is the HTTP server for the library application.
type Server struct {
	lib     *service.Library
	dataDir string
	port    int
}

// New creates a new Server.
func New(lib *service.Library, dataDir string, port int) *Server {
	return &Server{lib: lib, dataDir: dataDir, port: port}
}

// Start configures routes and starts the HTTP server.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	api := newAPIHandlers(s.lib)
	web := newWebHandlers(s.lib, s.dataDir)

	// --- JSON API routes ---

	// Books
	mux.HandleFunc("/api/books", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.listBooks(w, r)
		case http.MethodPost:
			api.createBook(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/api/books/search", func(w http.ResponseWriter, r *http.Request) {
		api.searchBooks(w, r)
	})

	mux.HandleFunc("/api/books/", func(w http.ResponseWriter, r *http.Request) {
		segs := extractPathSegments(r.URL.Path)
		// /api/books/{id}
		// /api/books/{id}/status
		// /api/books/{id}/rating
		// /api/books/{id}/cover
		// /api/books/{id}/cover/refresh
		// /api/books/{id}/tags
		// /api/books/{id}/tags/{tid}

		if len(segs) < 3 {
			writeError(w, http.StatusNotFound, "not found")
			return
		}

		// Determine sub-resource
		subResource := ""
		if len(segs) >= 4 {
			subResource = segs[3]
		}
		subSub := ""
		if len(segs) >= 5 {
			subSub = segs[4]
		}

		switch {
		case subResource == "status" && r.Method == http.MethodPatch:
			api.updateBookStatus(w, r)
		case subResource == "rating" && r.Method == http.MethodPatch:
			api.updateBookRating(w, r)
		case subResource == "cover" && subSub == "refresh" && r.Method == http.MethodPost:
			api.refreshCover(w, r)
		case subResource == "cover" && r.Method == http.MethodPost:
			api.setCover(w, r)
		case subResource == "tags" && len(segs) >= 6 && r.Method == http.MethodDelete:
			api.untagBook(w, r)
		case subResource == "tags" && r.Method == http.MethodPost:
			api.tagBook(w, r)
		case subResource == "loans" && r.Method == http.MethodGet:
			api.getBookLoans(w, r)
		case subResource == "" && r.Method == http.MethodGet:
			api.getBook(w, r)
		case subResource == "" && r.Method == http.MethodPut:
			api.updateBook(w, r)
		case subResource == "" && r.Method == http.MethodDelete:
			api.deleteBook(w, r)
		default:
			writeError(w, http.StatusNotFound, "not found")
		}
	})

	// Shelves
	mux.HandleFunc("/api/shelves", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.listShelves(w, r)
		case http.MethodPost:
			api.createShelf(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/api/shelves/", func(w http.ResponseWriter, r *http.Request) {
		segs := extractPathSegments(r.URL.Path)
		subResource := ""
		if len(segs) >= 4 {
			subResource = segs[3]
		}

		switch {
		case subResource == "books" && r.Method == http.MethodPost:
			api.addBookToShelf(w, r)
		case subResource == "books" && r.Method == http.MethodDelete:
			api.removeBookFromShelf(w, r)
		case subResource == "" && r.Method == http.MethodGet:
			api.getShelf(w, r)
		case subResource == "" && r.Method == http.MethodPut:
			api.updateShelf(w, r)
		case subResource == "" && r.Method == http.MethodDelete:
			api.deleteShelf(w, r)
		default:
			writeError(w, http.StatusNotFound, "not found")
		}
	})

	// Tags
	mux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
		api.listTags(w, r)
	})

	// Stats
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		api.getStats(w, r)
	})

	// Import/Export
	mux.HandleFunc("/api/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		api.importData(w, r)
	})

	mux.HandleFunc("/api/export", func(w http.ResponseWriter, r *http.Request) {
		api.exportData(w, r)
	})

	// Loans
	mux.HandleFunc("/api/loans", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.listLoans(w, r)
		case http.MethodPost:
			api.createLoan(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/api/loans/", func(w http.ResponseWriter, r *http.Request) {
		segs := extractPathSegments(r.URL.Path)
		if len(segs) < 3 {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		switch r.Method {
		case http.MethodGet:
			api.getLoan(w, r)
		case http.MethodPatch:
			api.checkInLoan(w, r)
		case http.MethodDelete:
			api.deleteLoan(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	// ISBN Lookup
	mux.HandleFunc("/api/lookup/", func(w http.ResponseWriter, r *http.Request) {
		api.lookupISBN(w, r)
	})

	// --- Cover images ---
	coversDir := filepath.Join(s.dataDir, "covers")
	mux.Handle("/covers/", http.StripPrefix("/covers/", http.FileServer(http.Dir(coversDir))))

	// --- Web UI routes ---
	web.registerRoutes(mux)

	// --- Static files ---
	staticDir := filepath.Join(s.dataDir, "..", "web", "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// --- Service Worker (must be served from root for root scope) ---
	mux.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Service-Worker-Allowed", "/")
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(staticDir, "sw.js"))
	})

	// Apply middleware
	handler := logging(cors(mux))

	addr := fmt.Sprintf(":%d", s.port)
	printServerInfo(s.port)

	return http.ListenAndServe(addr, handler)
}

func printServerInfo(port int) {
	fmt.Printf("\n  Library Server running!\n\n")
	fmt.Printf("  Local:   http://localhost:%d\n", port)

	// Try to find LAN IP
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				fmt.Printf("  Network: http://%s:%d\n", ipnet.IP.String(), port)
			}
		}
	}

	fmt.Printf("\n  API:     http://localhost:%d/api/books\n", port)
	fmt.Printf("  Web UI:  http://localhost:%d/\n\n", port)

	log.SetPrefix("[library] ")
	log.Printf("Listening on :%d", port)
}
