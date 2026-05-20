package core

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService enthält die Geschäftslogik für Registrierung und Login.
// Er kennt nur das UserRepository-Interface, nicht die konkrete DB-Implementierung.
type AuthService struct {
	repo UserRepository
}

// NewAuthService erstellt einen neuen AuthService mit dem gegebenen Repository.
func NewAuthService(repo UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Register registriert einen neuen Benutzer.
// Das Passwort wird gehasht, bevor es gespeichert wird.
func (s *AuthService) Register(req RegisterRequest) (*User, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, errors.New("alle Felder müssen ausgefüllt sein")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hash),
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login prüft die Anmeldedaten und gibt ein JWT zurück.
func (s *AuthService) Login(req LoginRequest) (string, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return "", errors.New("benutzer nicht gefunden")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("falsches passwort")
	}

	return generateJWT(user)
}

// generateJWT erstellt ein signiertes JWT für den gegebenen Benutzer.
func generateJWT(user *User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT prüft ein JWT und gibt die Claims zurück.
func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("ungültige Signaturmethode")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("ungültiger token")
	}
	return claims, nil
}
