package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"shop-service/internal/handler"

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

	if err := seed(db); err != nil {
		log.Fatalf("Seed fehlgeschlagen: %v", err)
	}

	// ── Router ────────────────────────────────────────────────────────
	router := http.NewServeMux()
	h := handler.NewShopHandler(db)

	router.HandleFunc("/shop", h.GetShop) // GET  /shop?name=&category=&min_price=&max_price=
	router.HandleFunc("/cart", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetCart(w, r) // GET  /cart
		case http.MethodPost:
			h.AddToCart(w, r) // POST /cart
		case http.MethodDelete:
			h.ClearCart(w, r) // DELETE /cart  → ganzen Warenkorb leeren
		default:
			http.Error(w, "methode nicht erlaubt", http.StatusMethodNotAllowed)
		}
	})
	router.HandleFunc("/cart/item/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			h.RemoveFromCart(w, r)
		case http.MethodPatch:
			h.UpdateCartItem(w, r)
		default:
			http.Error(w, "methode nicht erlaubt", http.StatusMethodNotAllowed)
		}
	})

	port := getEnv("PORT", "8082")
	log.Printf("Shop-Service läuft auf http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func connectDB() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "shopuser"),
		getEnv("DB_PASSWORD", "shoppass"),
		getEnv("DB_NAME", "shopdb"),
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
		CREATE TABLE IF NOT EXISTS shop_items (
			id       SERIAL PRIMARY KEY,
			name     VARCHAR(200) UNIQUE NOT NULL,
			category VARCHAR(100) NOT NULL,
			price    NUMERIC(10,2) NOT NULL
		);

		CREATE TABLE IF NOT EXISTS cart_items (
			id           SERIAL PRIMARY KEY,
			user_id      INT NOT NULL,
			shop_item_id INT NOT NULL REFERENCES shop_items(id),
			quantity     INT NOT NULL DEFAULT 1,
			created_at   TIMESTAMP DEFAULT NOW(),
			UNIQUE(user_id, shop_item_id)
		);
	`)
	return err
}

// seed befüllt die shop_items-Tabelle beim ersten Start mit vordefinierten Artikeln.
// ON CONFLICT DO NOTHING stellt sicher, dass beim Neustart keine Duplikate entstehen.
func seed(db *sql.DB) error {
	items := []struct {
		name     string
		category string
		price    float64
	}{
		{"Laptop", "Elektronik", 999.99},
		{"Smartphone", "Elektronik", 699.99},
		{"Kopfhörer", "Elektronik", 79.99},
		{"Tastatur", "Elektronik", 49.99},
		{"Maus", "Elektronik", 29.99},
		{"Monitor", "Elektronik", 299.99},
		{"Go Programmierbuch", "Bücher", 34.99},
		{"Clean Code", "Bücher", 29.99},
		{"Rucksack", "Zubehör", 59.99},
		{"USB-C Hub", "Zubehör", 39.99},
	}

	for _, item := range items {
		_, err := db.Exec(
			`INSERT INTO shop_items (name, category, price)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (name) DO NOTHING`,
			item.name, item.category, item.price,
		)
		if err != nil {
			return err
		}
	}
	log.Printf("Shop-Items geladen.")
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
