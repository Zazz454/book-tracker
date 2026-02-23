# My Library

A personal library catalog with a Go REST API server, web UI (PWA), and CLI client.

## Features

- **Book Management** - Add, edit, delete, rate, review, and track reading status
- **Cover Art** - Automatically fetched from Open Library and Google Books APIs, stored locally
- **Custom Shelves & Tags** - Organize books into collections with flexible labels
- **Full-Text Search** - SQLite FTS5 powered search across titles, authors, and notes
- **ISBN Lookup** - Auto-populate book metadata by scanning or entering an ISBN
- **Barcode Scanner** - Scan ISBN barcodes with your phone camera (Chrome Android)
- **Import/Export** - CSV and JSON import/export, including Goodreads CSV format
- **Reading Statistics** - Track books read, pages, genres, ratings, and reading pace
- **PWA** - Install as an app on Android, works offline after first load
- **Auto Light/Dark Theme** - Follows system preference with manual override
- **CLI Client** - Full-featured command-line interface for power users

## Quick Start

### Docker (Recommended)

```bash
docker compose up -d
# Open http://localhost:8080
```

### From Source

```bash
# Requires Go 1.22+
make build
make run
# Open http://localhost:8080
```

### CLI

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

# See all commands
./bin/library --help
```

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

## Docker

```bash
# Build and run
docker compose up -d

# View logs
docker compose logs -f library

# Stop
docker compose down

# With nginx reverse proxy (production)
docker compose --profile production up -d
```

Data is persisted in `./data/` (mounted as a volume).

## API

All endpoints are under `/api/`. See [PLAN.md](PLAN.md) for the full endpoint reference.

Quick examples:

```bash
# Add a book
curl -X POST http://localhost:8080/api/books \
  -H "Content-Type: application/json" \
  -d '{"title":"Dune","author":"Frank Herbert","genre":"Science Fiction"}'

# List books
curl http://localhost:8080/api/books?status=reading&sort=title

# Search
curl http://localhost:8080/api/books/search?q=dune

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
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## License

MIT
