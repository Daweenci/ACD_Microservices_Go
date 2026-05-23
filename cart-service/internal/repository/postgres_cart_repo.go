package repository

import (
	"cart-service/internal/core"
	"database/sql"
)

type PostgresCartRepository struct {
	db *sql.DB
}

func NewPostgresCartRepository(db *sql.DB) *PostgresCartRepository {
	return &PostgresCartRepository{db: db}
}

func (r *PostgresCartRepository) GetByUserID(userID int) ([]core.CartItem, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, item_id, quantity, created_at
		 FROM cart_items WHERE user_id = $1 ORDER BY created_at`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []core.CartItem
	for rows.Next() {
		var item core.CartItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.ItemID, &item.Quantity, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []core.CartItem{}
	}
	return items, nil
}

func (r *PostgresCartRepository) GetEntry(userID, itemID int) (*core.CartItem, error) {
	item := &core.CartItem{}
	err := r.db.QueryRow(
		`SELECT id, user_id, item_id, quantity, created_at
		 FROM cart_items WHERE user_id = $1 AND item_id = $2`,
		userID, itemID,
	).Scan(&item.ID, &item.UserID, &item.ItemID, &item.Quantity, &item.CreatedAt)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresCartRepository) Add(userID, itemID, quantity int) error {
	_, err := r.db.Exec(
		`INSERT INTO cart_items (user_id, item_id, quantity) VALUES ($1, $2, $3)`,
		userID, itemID, quantity,
	)
	return err
}

func (r *PostgresCartRepository) UpdateQuantity(userID, itemID, newQuantity int) error {
	_, err := r.db.Exec(
		`UPDATE cart_items SET quantity = $1 WHERE user_id = $2 AND item_id = $3`,
		newQuantity, userID, itemID,
	)
	return err
}

func (r *PostgresCartRepository) RemoveEntry(userID, itemID int) error {
	_, err := r.db.Exec(
		`DELETE FROM cart_items WHERE user_id = $1 AND item_id = $2`,
		userID, itemID,
	)
	return err
}

func (r *PostgresCartRepository) ClearCart(userID int) error {
	_, err := r.db.Exec(`DELETE FROM cart_items WHERE user_id = $1`, userID)
	return err
}
