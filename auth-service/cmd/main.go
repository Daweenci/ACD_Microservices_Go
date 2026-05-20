package main

import (
	"auth-service/internal/handler"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Datenbankverbindung aufbauen
	db, err := connectDB()
	if err != nil {
		log.Fatalf("Konnte keine DB-Verbindung herstellen: %v", err)
	}
	defer db.Close()

	// Tabellen anlegen (Migration)
	if err := migrate(db); err != nil {
		log.Fatalf("Migration fehlgeschlagen: %v", err)
	}

	// Router mit Endpoints registrieren
	router := http.NewServeMux()
	h := handler.NewAuthHandler(db)

	router.HandleFunc("/register", h.Register)      // POST
	router.HandleFunc("/login", h.Login)            // POST
	router.HandleFunc("/validate", h.ValidateToken) // GET

	port := getEnv("PORT", "8081")
	log.Printf("Auth-Service läuft auf http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func connectDB() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "authuser"),
		getEnv("DB_PASSWORD", "authpass"),
		getEnv("DB_NAME", "authdb"),
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	log.Println("Datenbankverbindung erfolgreich.")
	return db, nil
}

func migrate(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id         SERIAL PRIMARY KEY,
		username   VARCHAR(100) UNIQUE NOT NULL,
		email      VARCHAR(150) UNIQUE NOT NULL,
		password   TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);`
	_, err := db.Exec(query)
	return err
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
