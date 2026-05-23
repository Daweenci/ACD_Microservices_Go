package handler

import (
	"cart-service/internal/core"
	"cart-service/internal/repository"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type CartHandler struct {
	service *core.CartService
}

func NewCartHandler(db *sql.DB) *CartHandler {
	repo := repository.NewPostgresCartRepository(db)
	service := core.NewCartService(repo)
	return &CartHandler{service: service}
}

// extractUserID liest den JWT aus dem Authorization-Header und validiert ihn lokal.
// Kein Netzwerkaufruf – Validierung geschieht direkt anhand des gemeinsamen Secrets.
func extractUserID(r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return 0, &httpError{msg: "kein token angegeben", status: http.StatusUnauthorized}
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := core.ValidateJWT(tokenStr)
	if err != nil {
		return 0, &httpError{msg: "ungültiger token", status: http.StatusUnauthorized}
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, &httpError{msg: "user_id nicht im token", status: http.StatusUnauthorized}
	}
	return int(userIDFloat), nil
}

// ── GET /cart ─────────────────────────────────────────────────────────

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	items, err := h.service.GetCart(userID)
	if err != nil {
		writeError(w, "fehler beim laden des warenkorbs", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// ── POST /cart ────────────────────────────────────────────────────────
// Body: { "item_id": 42, "quantity": 2 }

func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		writeErr(w, err)
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

// ── PATCH /cart/item/{id} ─────────────────────────────────────────────
// Body: { "delta": 1 } oder { "delta": -2 }
// Fällt quantity auf 0 oder darunter → Eintrag wird automatisch gelöscht.

func (h *CartHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	itemID, err := extractItemID(r)
	if err != nil {
		writeError(w, "ungültige item_id im pfad", http.StatusBadRequest)
		return
	}
	var req core.UpdateCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ungültiges JSON", http.StatusBadRequest)
		return
	}
	if err := h.service.UpdateCartItem(userID, itemID, req.Delta); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "warenkorb aktualisiert"})
}

// ── DELETE /cart/item/{id} ────────────────────────────────────────────

func (h *CartHandler) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	itemID, err := extractItemID(r)
	if err != nil {
		writeError(w, "ungültige item_id im pfad", http.StatusBadRequest)
		return
	}
	if err := h.service.RemoveFromCart(userID, itemID); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "item entfernt"})
}

// ── DELETE /cart ──────────────────────────────────────────────────────

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	if err := h.service.ClearCart(userID); err != nil {
		writeError(w, "fehler beim leeren des warenkorbs", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "warenkorb geleert"})
}

// ── Hilfsfunktionen ───────────────────────────────────────────────────

type httpError struct {
	msg    string
	status int
}

func (e *httpError) Error() string { return e.msg }

func extractItemID(r *http.Request) (int, error) {
	parts := strings.Split(r.URL.Path, "/")
	return strconv.Atoi(parts[len(parts)-1])
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeErr(w http.ResponseWriter, err error) {
	if he, ok := err.(*httpError); ok {
		writeError(w, he.msg, he.status)
		return
	}
	writeError(w, err.Error(), http.StatusInternalServerError)
}
