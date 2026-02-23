package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/user/library/internal/models"
	"github.com/user/library/internal/service"
)

type webHandlers struct {
	lib       *service.Library
	templates map[string]*template.Template
	dataDir   string
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
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
			return strings.Repeat("*", n)
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
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 2, 2006")
		},
		"formatDatePtr": func(t *time.Time) string {
			if t == nil {
				return ""
			}
			return t.Format("Jan 2, 2006")
		},
		"amazonURL": func(title, author string) string {
			q := url.QueryEscape(title + " " + author)
			return "https://www.amazon.com/s?k=" + q + "&i=stripbooks"
		},
		"worldcatURL": func(title, author string) string {
			q := url.QueryEscape(title + " " + author)
			return "https://www.worldcat.org/search?q=" + q
		},
		"openlibrarySearchURL": func(title, author string) string {
			q := url.QueryEscape(title + " " + author)
			return "https://openlibrary.org/search?q=" + q
		},
		"loanTypeLabel": func(t string) string {
			if t == "borrowed" {
				return "Borrowed"
			}
			return "Lent"
		},
		"daysUntilDue": func(due *time.Time) int {
			if due == nil {
				return 0
			}
			return int(time.Until(*due).Hours() / 24)
		},
	}
}

func newWebHandlers(lib *service.Library, dataDir string) *webHandlers {
	templatesDir := filepath.Join(dataDir, "..", "web", "templates")
	layoutFile := filepath.Join(templatesDir, "layout.html")

	// Parse each page template individually with the layout so that
	// block definitions (content, title, scripts) don't collide.
	pages := []string{
		"index.html",
		"books.html",
		"book_detail.html",
		"book_form.html",
		"shelves.html",
		"shelf_detail.html",
		"stats.html",
		"scan.html",
		"loans.html",
		"loan_form.html",
		"offline.html",
		"login.html",
		"register.html",
		"account.html",
		"admin_users.html",
	}

	templates := make(map[string]*template.Template)
	for _, page := range pages {
		t, err := template.New(page).Funcs(templateFuncs()).ParseFiles(
			layoutFile,
			filepath.Join(templatesDir, page),
		)
		if err != nil {
			log.Fatalf("Failed to parse template %s: %v", page, err)
		}
		templates[page] = t
	}

	return &webHandlers{lib: lib, templates: templates, dataDir: dataDir}
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
	mux.HandleFunc("/offline", h.offlinePage)
	mux.HandleFunc("/loans", h.loanList)
	mux.HandleFunc("/loans/new", h.loanForm)
	mux.HandleFunc("/shelves", h.shelfList)
	mux.HandleFunc("/shelves/", h.shelfDetail)
	mux.HandleFunc("/stats", h.statsPage)
	mux.HandleFunc("/scan", h.scanPage)
	mux.HandleFunc("/login", h.loginPage)
	mux.HandleFunc("/register", h.registerPage)
	mux.HandleFunc("/account", h.accountPage)
	mux.HandleFunc("/admin/users", h.adminUsersPage)
}

func (h *webHandlers) render(w http.ResponseWriter, name string, data interface{}) {
	t, ok := h.templates[name]
	if !ok {
		log.Printf("template %q not found", name)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("template error (%s): %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// baseData returns a map with common template data including OverdueCount for nav badges
// and the currently authenticated user.
func (h *webHandlers) baseData(r *http.Request, page string) map[string]interface{} {
	overdueCount, _ := h.lib.GetOverdueCount()
	data := map[string]interface{}{
		"Page":         page,
		"OverdueCount": overdueCount,
	}
	if user := UserFromContext(r); user != nil {
		data["User"] = user
	}
	return data
}

// --- Auth page handlers ---

func (h *webHandlers) loginPage(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to home
	if user := UserFromContext(r); user != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	data := map[string]interface{}{
		"Page":  "login",
		"Error": r.URL.Query().Get("error"),
	}

	// Check if any users exist - if not, redirect to register
	count, _ := h.lib.CountUsers()
	if count == 0 {
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	h.render(w, "login.html", data)
}

func (h *webHandlers) registerPage(w http.ResponseWriter, r *http.Request) {
	// Only allow registration if no users exist (first user setup)
	// or if current user is admin
	isOpen := h.lib.IsRegistrationOpen()
	user := UserFromContext(r)

	if !isOpen && (user == nil || !user.IsAdmin) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	data := map[string]interface{}{
		"Page":      "register",
		"Error":     r.URL.Query().Get("error"),
		"IsFirstUser": isOpen,
	}
	if user != nil {
		data["User"] = user
	}

	h.render(w, "register.html", data)
}

func (h *webHandlers) accountPage(w http.ResponseWriter, r *http.Request) {
	data := h.baseData(r, "account")
	data["Success"] = r.URL.Query().Get("success")
	data["Error"] = r.URL.Query().Get("error")
	h.render(w, "account.html", data)
}

func (h *webHandlers) adminUsersPage(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r)
	if user == nil || !user.IsAdmin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	users, _ := h.lib.ListUsers()

	data := h.baseData(r, "admin")
	data["Users"] = users
	data["Success"] = r.URL.Query().Get("success")
	h.render(w, "admin_users.html", data)
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
	overdueResp, _ := h.lib.GetOverdueLoans()
	var overdueLoans []models.Loan
	if overdueResp != nil {
		overdueLoans = overdueResp.Loans
	}

	data := h.baseData(r, "home")
	data["Stats"] = stats
	data["Recent"] = recent
	data["Reading"] = reading
	data["OverdueLoans"] = overdueLoans
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

	data := h.baseData(r, "books")
	data["Books"] = result
	data["Params"] = params
	data["Shelves"] = shelves
	data["View"] = q.Get("view")
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
	loanHistory, _ := h.lib.GetBookLoanHistory(id)

	data := h.baseData(r, "books")
	data["Book"] = book
	data["AllShelves"] = shelves
	data["AllTags"] = tags
	data["LoanHistory"] = loanHistory
	h.render(w, "book_detail.html", data)
}

func (h *webHandlers) bookForm(w http.ResponseWriter, r *http.Request) {
	data := h.baseData(r, "books")
	data["Book"] = &models.Book{Status: "unread"}
	data["IsNew"] = true
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

	data := h.baseData(r, "books")
	data["Book"] = book
	data["IsNew"] = false
	h.render(w, "book_form.html", data)
}

func (h *webHandlers) shelfList(w http.ResponseWriter, r *http.Request) {
	shelves, err := h.lib.ListShelves()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := h.baseData(r, "shelves")
	data["Shelves"] = shelves
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

	data := h.baseData(r, "shelves")
	data["Shelf"] = shelf
	data["Books"] = books
	h.render(w, "shelf_detail.html", data)
}

func (h *webHandlers) statsPage(w http.ResponseWriter, r *http.Request) {
	stats, err := h.lib.GetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := h.baseData(r, "stats")
	data["Stats"] = stats
	h.render(w, "stats.html", data)
}

func (h *webHandlers) scanPage(w http.ResponseWriter, r *http.Request) {
	data := h.baseData(r, "scan")
	h.render(w, "scan.html", data)
}

func (h *webHandlers) offlinePage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"Page": ""}
	h.render(w, "offline.html", data)
}

func (h *webHandlers) loanList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := q.Get("status")
	if status == "" {
		status = "active"
	}
	params := models.LoanListParams{
		Status:   status,
		LoanType: q.Get("type"),
		Limit:    50,
	}
	if person := q.Get("person"); person != "" {
		params.PersonName = person
	}

	result, err := h.lib.ListLoans(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := h.baseData(r, "loans")
	data["Loans"] = result
	data["Status"] = status
	data["Type"] = q.Get("type")
	h.render(w, "loans.html", data)
}

func (h *webHandlers) loanForm(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	bookIDStr := q.Get("book_id")
	loanType := q.Get("type")
	if loanType == "" {
		loanType = "lent"
	}

	data := h.baseData(r, "loans")
	data["LoanType"] = loanType

	if bookIDStr != "" {
		bookID, err := strconv.ParseInt(bookIDStr, 10, 64)
		if err == nil {
			book, err := h.lib.GetBook(bookID)
			if err == nil {
				data["Book"] = book
			}
		}
	}

	// Provide a default due date 2 weeks from now
	data["DefaultDueDate"] = time.Now().AddDate(0, 0, 14).Format("2006-01-02")

	// List of books for the select dropdown (if no book pre-selected)
	if bookIDStr == "" {
		books, _ := h.lib.ListBooks(models.BookListParams{Limit: 500, Sort: "title", Order: "asc"})
		if books != nil {
			data["AllBooks"] = books.Books
		}
	}

	h.render(w, "loan_form.html", data)
}

// Ensure fmt is used
var _ = fmt.Sprintf
