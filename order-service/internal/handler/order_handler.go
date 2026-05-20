package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"order-service/internal/core"
	"order-service/internal/repository"
	"os"
	"strings"
)

// OrderHandler hält eine Referenz auf den OrderService.
type OrderHandler struct {
	service *core.OrderService
}

func NewOrderHandler(db *sql.DB) *OrderHandler {
	repo := repository.NewPostgresOrderRepository(db)
	service := core.NewOrderService(repo)
	return &OrderHandler{service: service}
}

// ── Response-Typen ────────────────────────────────────────────────────

type errorResponse struct {
	Error string `json:"error"`
}

// ── Middleware: JWT-Validierung über den Auth-Service ─────────────────

// extractUserID ruft den Auth-Service auf, um das JWT zu validieren,
// und gibt die user_id aus den Claims zurück.
// Das ist das Proxy-Pattern aus dem Artikel: der Order-Service kennt
// die JWT-Logik NICHT, er delegiert an den Auth-Service.
func extractUserID(r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return 0, fmt.Errorf("kein Token angegeben")
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
		return 0, fmt.Errorf("auth-service nicht erreichbar: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("ungültiger Token")
	}

	var claims map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return 0, err
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user_id nicht im Token")
	}
	return int(userIDFloat), nil
}

// ── Handler-Funktionen ────────────────────────────────────────────────

// CreateOrder verarbeitet POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "nur POST erlaubt", http.StatusMethodNotAllowed)
		return
	}

	userID, err := extractUserID(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req core.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ungültiges JSON", http.StatusBadRequest)
		return
	}

	order, err := h.service.CreateOrder(userID, req)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		panic(err)
	}
}

// GetOrders verarbeitet GET /orders
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "nur GET erlaubt", http.StatusMethodNotAllowed)
		return
	}

	userID, err := extractUserID(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	orders, err := h.service.GetOrdersByUser(userID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if orders == nil {
		orders = []core.Order{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		panic(err)
	}
}

func writeError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: msg})
}
