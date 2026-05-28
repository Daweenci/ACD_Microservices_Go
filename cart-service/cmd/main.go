package main

import (
	"cart-service/internal/handler"
	"database/sql"
	"fmt"
	"log"
	"net/http"
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
	h := handler.NewCartHandler(db)

	router.HandleFunc("/cart", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetCart(w, r)
		case http.MethodPost:
			h.AddToCart(w, r)
		case http.MethodDelete:
			h.ClearCart(w, r)
		default:
			http.Error(w, "methode nicht erlaubt", http.StatusMethodNotAllowed)
		}
	})
	router.HandleFunc("/cart/item/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			h.UpdateCartItem(w, r)
		case http.MethodDelete:
			h.RemoveFromCart(w, r)
		default:
			http.Error(w, "methode nicht erlaubt", http.StatusMethodNotAllowed)
		}
	})
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	port := getEnv("PORT", "8082")
	log.Printf("Cart-Service läuft auf http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func connectDB() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "cartuser"),
		getEnv("DB_PASSWORD", "cartpass"),
		getEnv("DB_NAME", "cartdb"),
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
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cart_items (
			id          SERIAL PRIMARY KEY,
			user_id     INT NOT NULL,
			item_id     INT NOT NULL,
			quantity    INT NOT NULL DEFAULT 1,
			created_at  TIMESTAMP DEFAULT NOW(),
			UNIQUE(user_id, item_id)
		);
	`)
	return err
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
