package core

import "time"

// User repräsentiert einen registrierten Benutzer.
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Passwort-Hash, nie im JSON
	CreatedAt time.Time `json:"created_at"`
}

// UserRepository definiert die Schnittstelle zur Datenpersistenz.
// Die Implementierung kennt die Geschäftslogik NICHT.
type UserRepository interface {
	Create(u *User) error
	FindByEmail(email string) (*User, error)
	FindByUsername(username string) (*User, error)
}

// RegisterRequest enthält die Eingabedaten für die Registrierung.
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest enthält die Eingabedaten für den Login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
