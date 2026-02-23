# Personal Library Catalog - Build Plan

## Architecture Overview

```
┌─────────────┐     HTTP/JSON     ┌──────────────┐     SQL     ┌──────────┐
│  CLI Client  │ ───────────────> │              │ ──────────> │          │
└─────────────┘                   │  REST API    │             │  SQLite  │
                                  │  Server (Go) │             │  (.db)   │
┌─────────────┐     HTTP/HTML     │              │ <────────── │          │
│  Web UI/PWA  │ <──────────────> │              │             │          │
└─────────────┘                   └──────────────┘             └──────────┘
```

**Stack:** Go, SQLite (via go-sqlite3), Cobra (CLI), html/template (web UI)
**Clients:** CLI + PWA (mobile-first web UI with auto light/dark theme)
**Database:** SQLite with FTS5 full-text search

---

## Implementation Phases

### Phase 0: Environment Setup
- [ ] Install Go, GCC, libsqlite3-dev
- [ ] Initialize Go module at /home/creator/projects/library/
- [ ] Install dependencies: go-sqlite3, cobra
- [ ] Create directory structure

### Phase 1: Data Layer
- [ ] Define models in `internal/models/models.go`
  - Book: ID, Title, Author, ISBN, Genre, Pages, Year, CoverPath, CoverSource, Status (unread/reading/finished/abandoned), Rating (0-5), Notes, StartedAt, FinishedAt, CreatedAt, UpdatedAt
  - Shelf: ID, Name, Description, CreatedAt
  - Tag: ID, Name
  - Request/response types for API
- [ ] Database connection in `internal/db/db.go` (WAL mode, foreign keys)
- [ ] Schema migrations in `internal/db/migrations.go`
  - books table
  - shelves table
  - book_shelves junction table
  - tags table
  - book_tags junction table
  - FTS5 virtual table on title, author, notes
- [ ] CRUD queries in `internal/db/queries.go`
  - Books: Create, GetByID, List (filters/sort/pagination), Update, Delete, UpdateStatus, UpdateRating, Search (FTS5)
  - Shelves: Create, List, GetByID, AddBook, RemoveBook, Delete
  - Tags: Create, List, TagBook, UntagBook
  - Stats: BooksPerMonth, BooksPerYear, GenreBreakdown, PagesRead, AverageRating

### Phase 2: Service Layer
- [ ] Business logic in `internal/service/library.go`
  - Orchestrates book creation + async cover fetching
  - Input validation (rating 1-5, status values, required fields)
  - Import/export logic (CSV/JSON parsing and generation)
  - Statistics computation

### Phase 3: REST API Server
- [ ] Routes in `internal/server/routes.go`
- [ ] JSON API handlers in `internal/server/handlers.go`
- [ ] Web (HTML) handlers in `internal/server/web_handlers.go`
- [ ] Middleware in `internal/server/middleware.go` (logging, CORS, content-type)
- [ ] Server entrypoint in `cmd/server/main.go` (--port, --data-dir flags, prints local IP)

**API Endpoints:**

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/books | List books (query: status, genre, shelf, tag, sort, q) |
| POST | /api/books | Add a book |
| GET | /api/books/:id | Get book detail |
| PUT | /api/books/:id | Update a book |
| DELETE | /api/books/:id | Delete a book |
| PATCH | /api/books/:id/status | Update reading status |
| PATCH | /api/books/:id/rating | Rate a book |
| POST | /api/books/:id/cover | Upload/set custom cover |
| POST | /api/books/:id/cover/refresh | Re-fetch cover from APIs |
| GET | /api/books/search?q= | Full-text search |
| GET | /api/shelves | List shelves |
| POST | /api/shelves | Create shelf |
| PUT | /api/shelves/:id | Update shelf |
| DELETE | /api/shelves/:id | Delete shelf |
| POST | /api/shelves/:id/books | Add book to shelf |
| DELETE | /api/shelves/:id/books/:bid | Remove book from shelf |
| GET | /api/tags | List tags |
| POST | /api/books/:id/tags | Tag a book |
| DELETE | /api/books/:id/tags/:tid | Remove tag |
| GET | /api/stats | Reading statistics |
| POST | /api/import | Import CSV/JSON |
| GET | /api/export?format=csv\|json | Export library |
| GET | /api/lookup/:isbn | Lookup ISBN metadata + cover |
| GET | /covers/:id.jpg | Serve cover image |

**Web Routes (HTML):**

| Path | Page |
|------|------|
| / | Dashboard / home |
| /books | Book grid/list |
| /books/:id | Book detail |
| /books/new | Add book form |
| /books/:id/edit | Edit book form |
| /shelves | Shelves list |
| /shelves/:id | Shelf detail |
| /stats | Statistics dashboard |
| /scan | Barcode scanner page |

### Phase 4: Cover Art System
- [ ] Cover orchestrator in `internal/covers/covers.go`
  - Tries Open Library first, then Google Books
  - Downloads image, saves to data/covers/{id}.jpg
  - Resizes to 300x450 max using pure Go image processing
  - Runs asynchronously (goroutine) so book creation isn't blocked
- [ ] Open Library client in `internal/covers/openlibrary.go`
  - Fetch by ISBN: covers.openlibrary.org/b/isbn/{isbn}-L.jpg
  - Fallback: search API by title+author
- [ ] Google Books client in `internal/covers/google.go`
  - Search by title+author, extract thumbnail URL
  - Free tier, no API key required for basic use

### Phase 5: Web UI (Mobile-First, PWA)
- [ ] Base layout template `web/templates/layout.html`
  - Nav, meta tags, manifest link, service worker registration
- [ ] Dashboard `web/templates/index.html`
  - Recent books, reading stats summary, quick-add
- [ ] Book list `web/templates/books.html`
  - Grid/list toggle, filter sidebar, sort dropdown
- [ ] Book detail `web/templates/book_detail.html`
  - Cover, metadata, status/rating controls, notes, shelves, tags
- [ ] Book form `web/templates/book_form.html`
  - Add/edit form with ISBN lookup button
- [ ] Shelves pages `web/templates/shelves.html`, `shelf_detail.html`
- [ ] Stats page `web/templates/stats.html`
  - Charts: reading pace, genre pie, rating distribution
- [ ] Scanner page `web/templates/scan.html`
  - Camera view for barcode scanning

**Auto-Switching Theme (Light/Dark):**
- [ ] CSS custom properties in `:root` and `@media (prefers-color-scheme: dark)`
  - Light: white/gray background, dark text, subtle shadows
  - Dark: dark navy/charcoal (#1a1a2e / #16213e), light text, subtle borders
- [ ] Manual override toggle in nav, stored in localStorage, synced to data-theme on html
- [ ] Smooth transition: `transition: background-color 0.3s, color 0.3s`

**Static Assets:**
- [ ] `web/static/style.css` - Mobile-first responsive CSS with auto theme
- [ ] `web/static/app.js` - SW registration, theme toggle, interactivity
- [ ] `web/static/scanner.js` - BarcodeDetector API wrapper, camera access
- [ ] `web/static/charts.js` - Lightweight canvas chart rendering
- [ ] `web/static/manifest.json` - PWA manifest
- [ ] `web/static/sw.js` - Service worker (cache app shell, covers; network-first for API)
- [ ] `web/static/icon-192.png`, `icon-512.png` - App icons

**Responsive Breakpoints:**
- Mobile (<640px): Single column, bottom nav
- Tablet (640-1024px): 2-3 column grid, sidebar filters
- Desktop (>1024px): 4-5 column grid, full sidebar

**PWA Features:**
- [ ] Service worker caching: app shell, CSS/JS, covers (cache-first), API (network-first)
- [ ] Offline fallback page
- [ ] Add to home screen support
- [ ] Share target registration

### Phase 6: CLI Client
- [ ] CLI entrypoint in `cmd/library/main.go` using Cobra
- [ ] Config file at ~/.library.yaml (server URL, default http://localhost:8080)
- [ ] --json flag on all commands for script-friendly output
- [ ] Tabular output for lists, detailed view for single items

**Commands:**

```
library serve                         # start the server
library add "Title" --author "Name"   # add a book
library add --isbn 978...            # add by ISBN (auto-lookup)
library list                          # list all books
library list --status reading         # filter by status
library list --shelf "Favorites"      # filter by shelf
library search "search terms"         # full-text search
library show <id>                     # book details
library edit <id> --title "New Title" # edit fields
library delete <id>                   # delete a book
library status <id> reading           # update status
library rate <id> 4                   # rate 1-5
library cover <id>                    # show cover info
library cover <id> --refresh          # re-fetch cover
library cover <id> --url <url>        # set custom cover
library shelves                       # list shelves
library shelves create "Name"         # create shelf
library shelves add <shelf> <book>    # add book to shelf
library tags                          # list tags
library tag <book-id> "fiction"       # tag a book
library stats                         # reading statistics
library import books.csv              # import from CSV/JSON
library export --format csv           # export to CSV/JSON
library lookup <isbn>                 # lookup ISBN metadata
library scan                          # open barcode scanner (web)
```

### Phase 7: ISBN Lookup + Barcode Scanner
- [ ] Open Library search client in `internal/openlibrary/client.go`
  - Search by ISBN: openlibrary.org/api/books?bibkeys=ISBN:{isbn}&format=json&jscmd=data
  - Returns: title, authors, publishers, publish date, page count, cover URLs, subjects
  - Auto-populates all fields when user provides ISBN
- [ ] Barcode scanner (web only)
  - BarcodeDetector API (native in Chrome Android)
  - Fallback to manual text input
  - On scan: calls /api/lookup/{isbn}, shows preview, user confirms to add

### Phase 8: Import/Export
- [ ] CSV import: column mapping to book fields, header row required
- [ ] JSON import: array of book objects
- [ ] CSV/JSON export: all books with metadata
- [ ] Goodreads CSV import: parse Goodreads export format specifically
- [ ] Bulk cover download on import (queued, rate-limited)

### Phase 9: Statistics Dashboard
- [ ] Books read per month (bar chart)
- [ ] Pages read over time (line chart)
- [ ] Genre distribution (pie/donut chart)
- [ ] Average rating given
- [ ] Reading speed (days per book)
- [ ] Longest/shortest books read
- [ ] Current reading streak
- [ ] All rendered with canvas elements, data from /api/stats

### Phase 10: Full-Text Search (FTS5)
- [ ] SQLite FTS5 virtual table indexing: title, author, notes, tags
- [ ] Prefix matching, phrase queries
- [ ] Rank results by relevance
- [ ] Highlighted snippets in search results
- [ ] Accessible from API and web UI search bar

---

## Database Schema

### books
| Column | Type | Notes |
|--------|------|-------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| title | TEXT | NOT NULL |
| author | TEXT | NOT NULL |
| isbn | TEXT | UNIQUE, nullable |
| genre | TEXT | |
| pages | INTEGER | |
| year | INTEGER | Publication year |
| cover_path | TEXT | Local path: covers/{id}.jpg |
| cover_source | TEXT | openlibrary, google, manual, or empty |
| status | TEXT | unread, reading, finished, abandoned (default: unread) |
| rating | INTEGER | 0-5 (0 = unrated) |
| notes | TEXT | Personal notes/review |
| started_at | DATETIME | |
| finished_at | DATETIME | |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP |
| updated_at | DATETIME | DEFAULT CURRENT_TIMESTAMP |

### shelves
| Column | Type | Notes |
|--------|------|-------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| name | TEXT | NOT NULL, UNIQUE |
| description | TEXT | |
| created_at | DATETIME | DEFAULT CURRENT_TIMESTAMP |

### book_shelves
| Column | Type | Notes |
|--------|------|-------|
| book_id | INTEGER | FK -> books.id ON DELETE CASCADE |
| shelf_id | INTEGER | FK -> shelves.id ON DELETE CASCADE |
| | | PRIMARY KEY (book_id, shelf_id) |

### tags
| Column | Type | Notes |
|--------|------|-------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| name | TEXT | NOT NULL, UNIQUE |

### book_tags
| Column | Type | Notes |
|--------|------|-------|
| book_id | INTEGER | FK -> books.id ON DELETE CASCADE |
| tag_id | INTEGER | FK -> tags.id ON DELETE CASCADE |
| | | PRIMARY KEY (book_id, tag_id) |

### books_fts (FTS5 virtual table)
- Indexes: title, author, notes
- Used for full-text search with ranking

---

## Project Structure

```
library/
├── PLAN.md
├── Makefile
├── go.mod
├── go.sum
├── cmd/
│   ├── server/
│   │   └── main.go
│   └── library/
│       └── main.go
├── internal/
│   ├── models/
│   │   └── models.go
│   ├── db/
│   │   ├── db.go
│   │   ├── migrations.go
│   │   └── queries.go
│   ├── service/
│   │   └── library.go
│   ├── server/
│   │   ├── routes.go
│   │   ├── handlers.go
│   │   ├── web_handlers.go
│   │   └── middleware.go
│   ├── covers/
│   │   ├── covers.go
│   │   ├── openlibrary.go
│   │   └── google.go
│   └── openlibrary/
│       └── client.go
├── web/
│   ├── templates/
│   │   ├── layout.html
│   │   ├── index.html
│   │   ├── books.html
│   │   ├── book_detail.html
│   │   ├── book_form.html
│   │   ├── shelves.html
│   │   ├── shelf_detail.html
│   │   ├── stats.html
│   │   └── scan.html
│   └── static/
│       ├── manifest.json
│       ├── sw.js
│       ├── style.css
│       ├── app.js
│       ├── scanner.js
│       ├── charts.js
│       ├── icon-192.png
│       └── icon-512.png
└── data/               (runtime, gitignored)
    ├── library.db
    └── covers/
```

---

## Network Access (for PWA on Android)

| Scenario | Approach |
|----------|----------|
| Same WiFi | Access via local IP (e.g., 192.168.1.50:8080) |
| Remote | Reverse proxy with HTTPS (Caddy/nginx + domain), or tunnel (Tailscale/Cloudflare Tunnel) |

Server prints local network address on startup for easy mobile discovery.

---

## Future: Native Android App

The REST API serves as the contract. Any native client talks to the same JSON endpoints.

Options when ready:
- Kotlin + Jetpack Compose (first-class Android)
- Flutter / React Native (cross-platform if iOS also desired)
- Server, database, API, and cover fetching all stay the same

The PWA should cover ~90% of mobile use cases. Go native only if you need background sync, widgets, or deeper OS integration.

---

## Dependencies

| Package | Purpose |
|---------|---------|
| github.com/mattn/go-sqlite3 | SQLite driver (CGO) |
| github.com/spf13/cobra | CLI framework |
| Standard library net/http | HTTP server |
| Standard library html/template | Server-side HTML rendering |
| Standard library image/* | Cover image processing |

## System Requirements

- Go 1.21+
- GCC (for CGO / go-sqlite3)
- libsqlite3-dev
