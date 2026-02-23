package server

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/user/library/internal/models"
	"github.com/user/library/internal/service"
)

type webHandlers struct {
	lib     *service.Library
	tmpl    *template.Template
	dataDir string
}

func newWebHandlers(lib *service.Library, dataDir string) *webHandlers {
	templatesDir := filepath.Join(dataDir, "..", "web", "templates")

	funcMap := template.FuncMap{
		"statusColor": func(s string) string {
			switch s {
			case "reading":
				return "status-reading"
			case "finished":
				return "status-finished"
			case "abandoned":
				return "status-abandoned"
			default:
				return "status-unread"
			}
		},
		"stars": func(n int) string {
			filled := strings.Repeat("*", n)
			empty := strings.Repeat("*", 5-n)
			_ = empty
			return filled
		},
		"ratingStars": func(n int) template.HTML {
			var sb strings.Builder
			for i := 1; i <= 5; i++ {
				if i <= n {
					sb.WriteString(`<span class="star filled">&#9733;</span>`)
				} else {
					sb.WriteString(`<span class="star">&#9734;</span>`)
				}
			}
			return template.HTML(sb.String())
		},
		"statusLabel": func(s string) string {
			switch s {
			case "reading":
				return "Reading"
			case "finished":
				return "Finished"
			case "abandoned":
				return "Abandoned"
			default:
				return "Unread"
			}
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i + 1
			}
			return s
		},
	}

	tmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob(filepath.Join(templatesDir, "*.html")))

	return &webHandlers{lib: lib, tmpl: tmpl, dataDir: dataDir}
}

func (h *webHandlers) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.dashboard)
	mux.HandleFunc("/books", h.bookList)
	mux.HandleFunc("/books/new", h.bookForm)
	mux.HandleFunc("/books/", func(w http.ResponseWriter, r *http.Request) {
		segs := extractPathSegments(r.URL.Path)
		if len(segs) >= 3 && segs[2] == "edit" {
			h.bookEdit(w, r)
			return
		}
		h.bookDetail(w, r)
	})
	mux.HandleFunc("/shelves", h.shelfList)
	mux.HandleFunc("/shelves/", h.shelfDetail)
	mux.HandleFunc("/stats", h.statsPage)
	mux.HandleFunc("/scan", h.scanPage)
}

func (h *webHandlers) render(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// --- Page handlers ---

func (h *webHandlers) dashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	stats, _ := h.lib.GetStats()
	recent, _ := h.lib.ListBooks(models.BookListParams{Sort: "created_at", Order: "desc", Limit: 8})
	reading, _ := h.lib.ListBooks(models.BookListParams{Status: "reading", Limit: 8})

	data := map[string]interface{}{
		"Page":     "home",
		"Stats":    stats,
		"Recent":   recent,
		"Reading":  reading,
	}
	h.render(w, "index.html", data)
}

func (h *webHandlers) bookList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	params := models.BookListParams{
		Status:  q.Get("status"),
		Genre:   q.Get("genre"),
		TagName: q.Get("tag"),
		Sort:    q.Get("sort"),
		Order:   q.Get("order"),
		Query:   q.Get("q"),
	}
	if lim := q.Get("limit"); lim != "" {
		params.Limit, _ = strconv.Atoi(lim)
	}
	if params.Limit == 0 {
		params.Limit = 48
	}
	if off := q.Get("offset"); off != "" {
		params.Offset, _ = strconv.Atoi(off)
	}

	var result *models.BookListResponse
	var err error
	if q.Get("q") != "" {
		result, err = h.lib.SearchBooks(q.Get("q"), params.Limit, params.Offset)
	} else {
		result, err = h.lib.ListBooks(params)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shelves, _ := h.lib.ListShelves()

	data := map[string]interface{}{
		"Page":    "books",
		"Books":   result,
		"Params":  params,
		"Shelves": shelves,
		"View":    q.Get("view"),
	}
	h.render(w, "books.html", data)
}

func (h *webHandlers) bookDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	book, err := h.lib.GetBook(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	shelves, _ := h.lib.ListShelves()
	tags, _ := h.lib.ListTags()

	data := map[string]interface{}{
		"Page":       "books",
		"Book":       book,
		"AllShelves": shelves,
		"AllTags":    tags,
	}
	h.render(w, "book_detail.html", data)
}

func (h *webHandlers) bookForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Page":   "books",
		"Book":   &models.Book{Status: "unread"},
		"IsNew":  true,
	}
	h.render(w, "book_form.html", data)
}

func (h *webHandlers) bookEdit(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	book, err := h.lib.GetBook(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Page":  "books",
		"Book":  book,
		"IsNew": false,
	}
	h.render(w, "book_form.html", data)
}

func (h *webHandlers) shelfList(w http.ResponseWriter, r *http.Request) {
	shelves, err := h.lib.ListShelves()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Page":    "shelves",
		"Shelves": shelves,
	}
	h.render(w, "shelves.html", data)
}

func (h *webHandlers) shelfDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	shelf, err := h.lib.GetShelf(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	books, _ := h.lib.GetShelfBooks(id)

	data := map[string]interface{}{
		"Page":  "shelves",
		"Shelf": shelf,
		"Books": books,
	}
	h.render(w, "shelf_detail.html", data)
}

func (h *webHandlers) statsPage(w http.ResponseWriter, r *http.Request) {
	stats, err := h.lib.GetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Page":  "stats",
		"Stats": stats,
	}
	h.render(w, "stats.html", data)
}

func (h *webHandlers) scanPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Page": "scan",
	}
	h.render(w, "scan.html", data)
}
