.PHONY: build run server cli clean

# Build both server and CLI
build: server cli

# Build the server binary
server:
	go build -o bin/library-server ./cmd/server

# Build the CLI binary
cli:
	go build -o bin/library ./cmd/library

# Run the server in development mode
run:
	go run ./cmd/server --data ./data --port 8080

# Run the server with auto-reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air --build.cmd "go build -o ./tmp/server ./cmd/server" --build.bin "./tmp/server --data ./data --port 8080"

# Clean build artifacts
clean:
	rm -rf bin/ tmp/

# Install dependencies
deps:
	go mod tidy

# Run tests
test:
	go test ./...
