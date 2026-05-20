package handler

import (
	"auth-service/internal/core"
	"auth-service/internal/repository"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

// AuthHandler hält eine Referenz auf den AuthService.
type AuthHandler struct {
	service *core.AuthService
}

// NewAuthHandler erstellt einen neuen Handler mit dem gegebenen AuthService.
func NewAuthHandler(db *sql.DB) *AuthHandler {
	repo := repository.NewPostgresUserRepository(db)
	service := core.NewAuthService(repo)
	return &AuthHandler{service: service}
}

// Response-Typen

type registerResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Message  string `json:"message"`
}

type loginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Handler-Funktionen

// Register verarbeitet POST /register
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
	if err := json.NewEncoder(w).Encode(registerResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Message:  "Registrierung erfolgreich",
	}); err != nil {
		panic(err)
	}
}

// Login verarbeitet POST /login
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
	if err := json.NewEncoder(w).Encode(loginResponse{
		Token:   token,
		Message: "Login erfolgreich",
	}); err != nil {
		panic(err)
	}
}

// ValidateToken verarbeitet GET /validate
// Wird vom Order-Service aufgerufen, um JWTs zu prüfen.
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "nur GET erlaubt", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		writeError(w, "kein Token angegeben", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := core.ValidateJWT(tokenStr)
	if err != nil {
		writeError(w, "ungültiger Token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(claims); err != nil {
		panic(err)
	}
}

// writeError schreibt eine JSON-Fehlermeldung.
func writeError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: msg})
}
