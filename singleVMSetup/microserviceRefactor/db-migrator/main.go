package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

const (
	maxRetries     = 30
	retryDelay     = 2 * time.Second
	migrationsPath = "/migrations"
)

func main() {

	dbHost := getEnv("DB_HOST", "postgres")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "maildb")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := waitForDB(dsn)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	log.Println("Database migration check completed")
}

func waitForDB(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				return db, nil
			}
		}
		log.Printf("Waiting for database... (attempt %d/%d)", i+1, maxRetries)
		time.Sleep(retryDelay)
	}
	return nil, fmt.Errorf("failed to connect to database after %d attempts: %v", maxRetries, err)
}

func runMigrations(db *sql.DB) error {

	goose.SetBaseFS(os.DirFS(migrationsPath))

	_, err := db.Exec("SELECT 1 FROM goose_db_version LIMIT 1")
	if err != nil {
		log.Println("First-time migration: creating goose version table")
		if err := goose.Create(db, migrationsPath, "", "sql"); err != nil {
			return fmt.Errorf("failed to initialize goose: %v", err)
		}
	}

	current, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current DB version: %v", err)
	}

	migrations, err := goose.CollectMigrations(migrationsPath, 0, goose.MaxVersion)
	if err != nil {
		return fmt.Errorf("failed to collect migrations: %v", err)
	}

	if len(migrations) == 0 {
		log.Println("No migrations found in", migrationsPath)
		return nil
	}

	latest := migrations[len(migrations)-1].Version
	if current >= latest {
		log.Printf("Database already up-to-date (version %d)", current)
		return nil
	}

	log.Printf("Current DB version: %d, Latest available: %d", current, latest)
	log.Printf("Applying %d pending migration(s)...", len(migrations))

	if err := goose.Up(db, migrationsPath); err != nil {
		return fmt.Errorf("migration failed: %v", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
