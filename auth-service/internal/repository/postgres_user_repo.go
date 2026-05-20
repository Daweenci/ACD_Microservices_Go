package repository

import (
	"auth-service/internal/core"
	"database/sql"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(u *core.User) error {
	query := `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRow(query, u.Username, u.Email, u.Password).
		Scan(&u.ID, &u.CreatedAt)
}

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

func (r *PostgresUserRepository) FindByID(id int) (*core.User, error) {
	u := &core.User{}
	query := `SELECT id, username, email, password, created_at FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *PostgresUserRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *PostgresUserRepository) UpdateUsername(id int, newUsername string) error {
	_, err := r.db.Exec(`UPDATE users SET username = $1 WHERE id = $2`, newUsername, id)
	return err
}
