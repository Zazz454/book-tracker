# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy source code (go mod tidy needs to see source files for golang.org/x/crypto)
COPY . .

# Download and tidy dependencies
RUN go mod tidy && go mod download

# Build the server binary (pure Go, no CGO needed)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o library-server ./cmd/server

# Build the CLI binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o library-cli ./cmd/library

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests to external APIs
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 library && \
    adduser -D -s /bin/sh -u 1000 -G library library

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/library-server .
COPY --from=builder /app/library-cli .

# Copy web assets
COPY --from=builder /app/web ./web

# Create data directory and set permissions
RUN mkdir -p /app/data && \
    chown -R library:library /app

# Switch to non-root user
USER library

# Expose port
EXPOSE 8080

# Set environment variables
ENV LIBRARY_DATA_DIR=/app/data

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/login || exit 1

# Run the server
CMD ["./library-server", "--port", "8080", "--data", "/app/data"]
