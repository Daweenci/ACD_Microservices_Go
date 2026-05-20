package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"shop-service/internal/core"
	"shop-service/internal/repository"
	"strconv"
	"strings"
)

type ShopHandler struct {
	service *core.ShopService
}

func NewShopHandler(db *sql.DB) *ShopHandler {
	repo := repository.NewPostgresShopRepository(db)
	service := core.NewShopService(repo)
	return &ShopHandler{service: service}
}

// ── JWT-Validierung via Auth-Service ─────────────────────────────────

func extractUserID(r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return 0, fmt.Errorf("kein token angegeben")
	}

	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://localhost:8081"
	}

	req, err := http.NewRequest(http.MethodGet, authURL+"/validate", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("auth-service nicht erreichbar")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("ungültiger token")
	}

	var claims map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return 0, err
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user_id nicht im token")
	}
	return int(userIDFloat), nil
}

// ── GET /shop ─────────────────────────────────────────────────────────
// Optionale Query-Parameter: name, category, min_price, max_price

func (h *ShopHandler) GetShop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "nur GET erlaubt", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	filter := core.ShopFilter{
		Name:     q.Get("name"),
		Category: q.Get("category"),
	}
	if v := q.Get("min_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MinPrice = f
		}
	}
	if v := q.Get("max_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filter.MaxPrice = f
		}
	}

	items, err := h.service.GetShopItems(filter)
	if err != nil {
		writeError(w, "fehler beim laden der shop-items", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// ── POST /cart ────────────────────────────────────────────────────────
// Body: { "shop_item_id": 1, "quantity": 2 }

func (h *ShopHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "nur POST erlaubt", http.StatusMethodNotAllowed)
		return
	}

	userID, err := extractUserID(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req core.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ungültiges JSON", http.StatusBadRequest)
		return
	}

	if err := h.service.AddToCart(userID, req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "item zum warenkorb hinzugefügt"})
}

// ── GET /cart ─────────────────────────────────────────────────────────

func (h *ShopHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "nur GET erlaubt", http.StatusMethodNotAllowed)
		return
	}

	userID, err := extractUserID(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	items, err := h.service.GetCart(userID)
	if err != nil {
		writeError(w, "fehler beim laden des warenkorbs", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// ── DELETE /cart/item/{id} ────────────────────────────────────────────
// Optionaler Query-Parameter: quantity
// Kein quantity → 1 abziehen
// quantity=0    → nichts tun
// quantity=X    → X abziehen, bei Überschreitung alles löschen

func (h *ShopHandler) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, "nur DELETE erlaubt", http.StatusMethodNotAllowed)
		return
	}

	userID, err := extractUserID(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// shop_item_id aus dem Pfad lesen: /cart/item/3
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		writeError(w, "shop_item_id fehlt im pfad", http.StatusBadRequest)
		return
	}
	shopItemID, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		writeError(w, "ungültige shop_item_id", http.StatusBadRequest)
		return
	}

	// quantity aus Query-Parameter
	quantityProvided := false
	quantity := 0
	if qStr := r.URL.Query().Get("quantity"); qStr != "" {
		quantityProvided = true
		quantity, err = strconv.Atoi(qStr)
		if err != nil || quantity < 0 {
			writeError(w, "ungültige quantity", http.StatusBadRequest)
			return
		}
	}

	if err := h.service.RemoveFromCart(userID, shopItemID, quantity, quantityProvided); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "warenkorb aktualisiert"})
}

// ── DELETE /cart ──────────────────────────────────────────────────────

func (h *ShopHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, "nur DELETE erlaubt", http.StatusMethodNotAllowed)
		return
	}

	userID, err := extractUserID(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := h.service.ClearCart(userID); err != nil {
		writeError(w, "fehler beim leeren des warenkorbs", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "warenkorb geleert"})
}

// ── Hilfsfunktionen ───────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *ShopHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeError(w, "nur PATCH erlaubt", http.StatusMethodNotAllowed)
		return
	}

	userID, err := extractUserID(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	shopItemID, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		writeError(w, "ungültige shop_item_id", http.StatusBadRequest)
		return
	}

	var body struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "ungültiges JSON", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateCartItemQuantity(userID, shopItemID, body.Quantity); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "warenkorb aktualisiert"})
}
