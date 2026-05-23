package handler

import (
	"auth-service/internal/core"
	"auth-service/internal/repository"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

type AuthHandler struct {
	service *core.AuthService
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	repo := repository.NewPostgresUserRepository(db)
	service := core.NewAuthService(repo)
	return &AuthHandler{service: service}
}

// ── Response-Typen ────────────────────────────────────────────────────

type errorResponse struct {
	Error string `json:"error"`
}

type messageResponse struct {
	Message string `json:"message"`
}

// ── Hilfsfunktion: JWT aus Header lesen und validieren ────────────────

func extractClaims(r *http.Request) (map[string]interface{}, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, nil
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	return core.ValidateJWT(tokenStr)
}

// ── Handler ───────────────────────────────────────────────────────────

// POST /register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "nur POST erlaubt", http.StatusMethodNotAllowed)
		return
	}
	var req core.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ungültiges JSON", http.StatusBadRequest)
		return
	}
	user, err := h.service.Register(req)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"message":  "registrierung erfolgreich",
	})
}

// POST /login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "nur POST erlaubt", http.StatusMethodNotAllowed)
		return
	}
	var req core.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ungültiges JSON", http.StatusBadRequest)
		return
	}
	token, err := h.service.Login(req)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":   token,
		"message": "login erfolgreich",
	})
}

// GET /validate  – wird vom Order-Service aufgerufen
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "nur GET erlaubt", http.StatusMethodNotAllowed)
		return
	}
	claims, err := extractClaims(r)
	if err != nil || claims == nil {
		writeError(w, "ungültiger oder fehlender token", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(claims)
}

// DELETE /user  – löscht den eingeloggten Benutzer und versucht seinen Warenkorb zu leeren (JWT required)
func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, "nur DELETE erlaubt", http.StatusMethodNotAllowed)
		return
	}
	claims, err := extractClaims(r)
	if err != nil || claims == nil {
		writeError(w, "authentifizierung erforderlich", http.StatusUnauthorized)
		return
	}
	userID := int(claims["user_id"].(float64))
	if err := h.service.DeleteUser(userID); err != nil {
		writeError(w, "benutzer konnte nicht gelöscht werden", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messageResponse{Message: "benutzer erfolgreich gelöscht"})
}

// PATCH /user/username  – ändert den Benutzernamen (JWT required)
func (h *AuthHandler) ChangeUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeError(w, "nur PATCH erlaubt", http.StatusMethodNotAllowed)
		return
	}
	claims, err := extractClaims(r)
	if err != nil || claims == nil {
		writeError(w, "authentifizierung erforderlich", http.StatusUnauthorized)
		return
	}
	var req core.ChangeUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "ungültiges JSON", http.StatusBadRequest)
		return
	}
	userID := int(claims["user_id"].(float64))
	if err := h.service.ChangeUsername(userID, req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messageResponse{Message: "benutzername erfolgreich geändert"})
}

func writeError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: msg})
}
