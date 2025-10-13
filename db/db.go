package db

import ( 
	"embed"
	"fmt"
	"log"
	"os"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed words.db
var embeddedDB embed.FS

// The public db instance
var DB *sql.DB

// initDB initializes the database connection

func InitDB() error {
	dbContent, err := embeddedDB.ReadFile("words.db")
	if err != nil {
		log.Fatalf("Error reading embedded database: %v", err)
	}

	// Create a temporary file to store the database content
	tempFile, err := os.CreateTemp("", "embedded-db-*.db")
	if err != nil {
		log.Fatalf("Error creating temporary file: %v", err)
	}
	defer tempFile.Close()

	_, err = tempFile.Write(dbContent)
	if err != nil {
		log.Fatalf("Error writing to temporary file: %v", err)
	}

	// Start up the database
	DB, err = sql.Open("sqlite3", tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Database connection initialized")
	return nil
}

