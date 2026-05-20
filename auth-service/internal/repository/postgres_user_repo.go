package repository

import (
	"auth-service/internal/core"
	"database/sql"
)

// PostgresUserRepository implementiert core.UserRepository mit PostgreSQL.
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository erstellt eine neue Instanz.
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Create speichert einen neuen Benutzer in der Datenbank.
func (r *PostgresUserRepository) Create(u *core.User) error {
	query := `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRow(query, u.Username, u.Email, u.Password).
		Scan(&u.ID, &u.CreatedAt)
}

// FindByEmail sucht einen Benutzer anhand seiner E-Mail-Adresse.
func (r *PostgresUserRepository) FindByEmail(email string) (*core.User, error) {
	u := &core.User{}
	query := `SELECT id, username, email, password, created_at FROM users WHERE email = $1`
	err := r.db.QueryRow(query, email).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// FindByUsername sucht einen Benutzer anhand seines Benutzernamens.
func (r *PostgresUserRepository) FindByUsername(username string) (*core.User, error) {
	u := &core.User{}
	query := `SELECT id, username, email, password, created_at FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}
