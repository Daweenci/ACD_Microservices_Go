package repository

import (
	"database/sql"
	"order-service/internal/core"
)

// PostgresOrderRepository implementiert core.OrderRepository.
type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(o *core.Order) error {
	query := `
		INSERT INTO orders (user_id, item, quantity)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRow(query, o.UserID, o.Item, o.Quantity).
		Scan(&o.ID, &o.CreatedAt)
}

func (r *PostgresOrderRepository) FindByUserID(userID int) ([]core.Order, error) {
	query := `SELECT id, user_id, item, quantity, created_at FROM orders WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []core.Order
	for rows.Next() {
		var o core.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Item, &o.Quantity, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}
