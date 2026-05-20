package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"order-service/internal/handler"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	db, err := connectDB()
	if err != nil {
		log.Fatalf("Konnte keine DB-Verbindung herstellen: %v", err)
	}
	defer db.Close()

	if err := migrate(db); err != nil {
		log.Fatalf("Migration fehlgeschlagen: %v", err)
	}

	router := http.NewServeMux()
	h := handler.NewOrderHandler(db)

	router.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateOrder(w, r)
		case http.MethodGet:
			h.GetOrders(w, r)
		default:
			http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		}
	})

	port := getEnv("PORT", "8082")
	log.Printf("Order-Service läuft auf http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func connectDB() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "orderuser"),
		getEnv("DB_PASSWORD", "orderpass"),
		getEnv("DB_NAME", "orderdb"),
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
	CREATE TABLE IF NOT EXISTS orders (
		id         SERIAL PRIMARY KEY,
		user_id    INT NOT NULL,
		item       VARCHAR(200) NOT NULL,
		quantity   INT NOT NULL DEFAULT 1,
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
