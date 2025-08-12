package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run migrate.go [up|down|create] [name]")
	}

	command := os.Args[1]
	
	switch command {
	case "up":
		runMigrationsUp()
	case "down":
		runMigrationsDown()
	case "create":
		if len(os.Args) < 3 {
			log.Fatal("Usage: go run migrate.go create <migration_name>")
		}
		createMigration(os.Args[2])
	default:
		log.Fatal("Unknown command. Use: up, down, or create")
	}
}

func getDBConnection() (*gorm.DB, error) {
	// Read from environment or use defaults
	dbHost := getEnv("MASTER_DB_HOST", "localhost")
	dbPort := getEnv("MASTER_DB_PORT", "5432")
	dbUser := getEnv("MASTER_DB_USER", "saas_user")
	dbPassword := getEnv("MASTER_DB_PASSWORD", "saas_password")
	dbName := getEnv("MASTER_DB_NAME", "master_db")
	sslMode := getEnv("MASTER_DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort, sslMode)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func runMigrationsUp() {
	log.Println("Running migrations up...")
	
	db, err := getDBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create migrations table if it doesn't exist
	db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL,
			executed_at TIMESTAMP DEFAULT NOW()
		)
	`)

	// Get list of migration files
	migrationFiles, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil {
		log.Fatal("Failed to read migration files:", err)
	}

	for _, file := range migrationFiles {
		filename := filepath.Base(file)
		
		// Check if migration already executed
		var count int64
		db.Raw("SELECT COUNT(*) FROM migrations WHERE filename = ?", filename).Scan(&count)
		
		if count > 0 {
			log.Printf("Migration %s already executed, skipping", filename)
			continue
		}

		// Read migration file
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatal("Failed to read migration file:", err)
		}

		// Execute migration
		if err := db.Exec(string(content)).Error; err != nil {
			log.Fatal("Failed to execute migration:", err)
		}

		// Record migration
		db.Exec("INSERT INTO migrations (filename) VALUES (?)", filename)
		log.Printf("Executed migration: %s", filename)
	}

	log.Println("All migrations completed successfully!")
}

func runMigrationsDown() {
	log.Println("Running migrations down...")
	
	db, err := getDBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get last executed migration
	var lastMigration string
	db.Raw("SELECT filename FROM migrations ORDER BY executed_at DESC LIMIT 1").Scan(&lastMigration)
	
	if lastMigration == "" {
		log.Println("No migrations to rollback")
		return
	}

	// Find corresponding down migration
	downFile := "../../migrations/" + lastMigration[:len(lastMigration)-6] + "down.sql"
	
	if _, err := os.Stat(downFile); os.IsNotExist(err) {
		log.Fatal("Down migration file not found:", downFile)
	}

	// Read and execute down migration
	content, err := os.ReadFile(downFile)
	if err != nil {
		log.Fatal("Failed to read down migration file:", err)
	}

	if err := db.Exec(string(content)).Error; err != nil {
		log.Fatal("Failed to execute down migration:", err)
	}

	// Remove migration record
	db.Exec("DELETE FROM migrations WHERE filename = ?", lastMigration)
	log.Printf("Rolled back migration: %s", lastMigration)
}

func createMigration(name string) {
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s", timestamp, name)
	
	// Create up migration file
	upFile := fmt.Sprintf("../../migrations/%s.up.sql", filename)
	upContent := fmt.Sprintf(`-- Migration: %s
-- Created: %s

-- Add your migration SQL here

`, name, time.Now().Format("2006-01-02 15:04:05"))
	
	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		log.Fatal("Failed to create up migration file:", err)
	}

	// Create down migration file
	downFile := fmt.Sprintf("../../migrations/%s.down.sql", filename)
	downContent := fmt.Sprintf(`-- Rollback migration: %s
-- Created: %s

-- Add your rollback SQL here

`, name, time.Now().Format("2006-01-02 15:04:05"))
	
	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		log.Fatal("Failed to create down migration file:", err)
	}

	log.Printf("Created migration files:")
	log.Printf("  - %s", upFile)
	log.Printf("  - %s", downFile)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}