package core

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRepository interface {
	Create(u *User) error
	FindByEmail(email string) (*User, error)
	FindByUsername(username string) (*User, error)
	FindByID(id int) (*User, error)
	Delete(id int) error
	UpdateUsername(id int, newUsername string) error
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	// Entweder email oder username
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type ChangeUsernameRequest struct {
	NewUsername string `json:"new_username"`
}
