package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/user/library/internal/db"
	"github.com/user/library/internal/server"
	"github.com/user/library/internal/service"
)

func main() {
	port := flag.Int("port", 8080, "Server port")
	dataDir := flag.String("data", "", "Data directory (default: ./data)")
	flag.Parse()

	if *dataDir == "" {
		exe, err := os.Executable()
		if err != nil {
			*dataDir = "./data"
		} else {
			*dataDir = filepath.Join(filepath.Dir(exe), "data")
		}
	}

	// Also check for DATA_DIR env var
	if envDir := os.Getenv("LIBRARY_DATA_DIR"); envDir != "" {
		*dataDir = envDir
	}

	database, err := db.Open(*dataDir)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	lib := service.NewLibrary(database)
	srv := server.New(lib, *dataDir, *port)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
