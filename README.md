# My Library

A personal library catalog with a Go REST API server, web UI (PWA), and CLI client.

## Features

- **Book Management** - Add, edit, delete, rate, review, and track reading status
- **Lending & Borrowing** - Check out books to friends, track who has what, overdue warnings
- **Cover Art** - Automatically fetched from Open Library and Google Books APIs, stored locally
- **Custom Shelves & Tags** - Organize books into collections with flexible labels
- **Full-Text Search** - SQLite FTS5 powered search across titles, authors, and notes
- **ISBN Lookup** - Auto-populate book metadata by scanning or entering an ISBN
- **Barcode Scanner** - Scan ISBN barcodes with your phone camera (Chrome Android)
- **Import/Export** - CSV and JSON import/export, including Goodreads CSV format
- **Reading Statistics** - Track books read, pages, genres, ratings, and reading pace
- **External Links** - Amazon buy links, WorldCat and Open Library search from book detail
- **PWA** - Install as an app on Android, works offline after first load
- **Auto Light/Dark Theme** - Follows system preference with manual override
- **CLI Client** - Full-featured command-line interface for power users

## Quick Start

### Docker (Recommended)

Pull the pre-built image from GitHub Container Registry and run:

```bash
docker compose up -d
# Open http://localhost:8080
```

That's it. Data is persisted in a Docker volume (`library_data`).

### Build from Source

```bash
# Requires Go 1.24+
make build
make run
# Open http://localhost:8080
```

Or build the Docker image locally:

```bash
docker build -t book-tracker .
docker run -d -p 8080:8080 -v library_data:/app/data book-tracker
```

## Docker

The `docker-compose.yml` pulls `ghcr.io/zazz454/book-tracker:latest`, which is automatically built on every push to `main`.

```bash
# Start
docker compose up -d

# View logs
docker compose logs -f library

# Stop
docker compose down

# Update to latest image
docker compose pull && docker compose up -d

# With nginx reverse proxy (production)
docker compose --profile production up -d
```

Data is stored in a named Docker volume (`library_data`). To back it up:

```bash
docker compose cp library:/app/data ./backup
```

## CLI

```bash
# Build the CLI
make cli

# Add a book
./bin/library add "Dune" --author "Frank Herbert" --genre "Science Fiction"

# Add by ISBN (auto-lookup)
./bin/library add --isbn 9780441172719

# List books
./bin/library list
./bin/library list --status reading

# Search
./bin/library search "science fiction"

# Lend a book
./bin/library checkout 1 --person "Alice" --due 2026-04-01

# Check who has your books
./bin/library loans

# Return a book
./bin/library checkin 1

# See all commands
./bin/library --help
```

## Lending & Borrowing

Track books you've lent out or borrowed from others:

- **Lend a book** - Record who you lent it to, with optional due date
- **Borrow a book** - Track where you borrowed it from
- **Overdue warnings** - Dashboard alerts and nav badges for overdue items
- **Loan history** - Full history on each book's detail page

From the web UI, use the Loans tab in the bottom nav or the lend/borrow buttons on any book's detail page.

## Architecture

```
Client (CLI/Web/PWA)  -->  REST API (Go)  -->  SQLite (+ FTS5)
                                |
                          Cover Fetcher
                          (Open Library / Google Books)
```

- **Server**: Go standard library `net/http`, no frameworks
- **Database**: SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- **CLI**: Cobra
- **Web UI**: Server-side rendered Go `html/template`
- **PWA**: Service worker, manifest, offline support
- **CI/CD**: GitHub Actions builds and pushes Docker image to `ghcr.io` on every push to `main`

## API

All endpoints are under `/api/`. See [PLAN.md](PLAN.md) for the full endpoint reference.

```bash
# Add a book
curl -X POST http://localhost:8080/api/books \
  -H "Content-Type: application/json" \
  -d '{"title":"Dune","author":"Frank Herbert","genre":"Science Fiction"}'

# List books
curl http://localhost:8080/api/books?status=reading&sort=title

# Search
curl http://localhost:8080/api/books/search?q=dune

# Lend a book
curl -X POST http://localhost:8080/api/loans \
  -H "Content-Type: application/json" \
  -d '{"book_id":1,"loan_type":"lent","person_name":"Alice","due_date":"2026-04-01"}'

# List active loans
curl http://localhost:8080/api/loans?status=active

# Return a book
curl -X PATCH http://localhost:8080/api/loans/1 \
  -H "Content-Type: application/json" -d '{}'

# ISBN lookup
curl http://localhost:8080/api/lookup/9780441172719

# Export
curl http://localhost:8080/api/export?format=json > library.json
```

## Mobile (PWA)

1. Open `http://<server-ip>:8080` in Chrome on Android
2. Tap the "Add to Home Screen" prompt
3. The app opens in standalone mode (no browser chrome)
4. Works offline for previously loaded data

To find your server's IP, check the output when the server starts -- it prints its local network addresses.

## Project Structure

```
library/
├── cmd/
│   ├── server/main.go        # Server entrypoint
│   └── library/main.go       # CLI entrypoint
├── internal/
│   ├── models/models.go      # Data types
│   ├── db/                   # SQLite database layer
│   ├── service/library.go    # Business logic
│   ├── server/               # HTTP handlers and routes
│   ├── covers/               # Cover art fetching
│   └── openlibrary/          # ISBN lookup client
├── web/
│   ├── templates/            # HTML templates
│   └── static/               # CSS, JS, PWA assets
├── .github/workflows/        # CI/CD (Docker build + GHCR push)
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## License

MIT
