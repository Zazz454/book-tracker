package main

import (
	"flag"
	"fmt"
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
	adminPassword := flag.String("admin-password", "", "Default admin password if no users exist (default: 'admin')")
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

	// Ensure default admin exists
	defaultPassword := *adminPassword
	if defaultPassword == "" {
		defaultPassword = "admin"
	}
	created, err := lib.EnsureDefaultAdmin(defaultPassword)
	if err != nil {
		log.Fatalf("Failed to ensure default admin: %v", err)
	}
	if created {
		fmt.Printf("\nDefault admin account created:\n")
		fmt.Printf("  Username: admin\n")
		fmt.Printf("  Password: %s\n", defaultPassword)
		fmt.Printf("  Please log in and change the password!\n\n")
	}

	srv := server.New(lib, *dataDir, *port)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}