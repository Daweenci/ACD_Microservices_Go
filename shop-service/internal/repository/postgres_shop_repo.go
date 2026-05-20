package repository

import (
	"database/sql"
	"fmt"
	"shop-service/internal/core"
)

type PostgresShopRepository struct {
	db *sql.DB
}

func NewPostgresShopRepository(db *sql.DB) *PostgresShopRepository {
	return &PostgresShopRepository{db: db}
}

// ── Shop ──────────────────────────────────────────────────────────────

// GetAllItems gibt alle Shop-Items zurück, optional gefiltert.
func (r *PostgresShopRepository) GetAllItems(filter core.ShopFilter) ([]core.ShopItem, error) {
	query := `SELECT id, name, category, price FROM shop_items WHERE 1=1`
	args := []interface{}{}
	i := 1

	if filter.Name != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", i)
		args = append(args, "%"+filter.Name+"%")
		i++
	}
	if filter.Category != "" {
		query += fmt.Sprintf(" AND category ILIKE $%d", i)
		args = append(args, "%"+filter.Category+"%")
		i++
	}
	if filter.MinPrice > 0 {
		query += fmt.Sprintf(" AND price >= $%d", i)
		args = append(args, filter.MinPrice)
		i++
	}
	if filter.MaxPrice > 0 {
		query += fmt.Sprintf(" AND price <= $%d", i)
		args = append(args, filter.MaxPrice)
		i++
	}
	query += " ORDER BY id"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []core.ShopItem
	for rows.Next() {
		var item core.ShopItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []core.ShopItem{}
	}
	return items, nil
}

// GetItemByID gibt ein einzelnes Shop-Item zurück.
func (r *PostgresShopRepository) GetItemByID(id int) (*core.ShopItem, error) {
	item := &core.ShopItem{}
	err := r.db.QueryRow(
		`SELECT id, name, category, price FROM shop_items WHERE id = $1`, id,
	).Scan(&item.ID, &item.Name, &item.Category, &item.Price)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// ── Warenkorb ─────────────────────────────────────────────────────────

// GetCartByUserID gibt den Warenkorb eines Benutzers zurück.
// Name kommt via JOIN aus shop_items – Preis wird NICHT mitgegeben.
func (r *PostgresShopRepository) GetCartByUserID(userID int) ([]core.CartItem, error) {
	query := `
		SELECT c.id, c.user_id, c.shop_item_id, s.name, c.quantity, c.created_at
		FROM cart_items c
		JOIN shop_items s ON s.id = c.shop_item_id
		WHERE c.user_id = $1
		ORDER BY c.created_at`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []core.CartItem
	for rows.Next() {
		var item core.CartItem
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.ShopItemID,
			&item.Name, &item.Quantity, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []core.CartItem{}
	}
	return items, nil
}

// GetCartEntry gibt einen einzelnen Warenkorbeintrag zurück (user + shop_item).
func (r *PostgresShopRepository) GetCartEntry(userID, shopItemID int) (*core.CartItem, error) {
	item := &core.CartItem{}
	err := r.db.QueryRow(
		`SELECT id, user_id, shop_item_id, quantity, created_at
		 FROM cart_items WHERE user_id = $1 AND shop_item_id = $2`,
		userID, shopItemID,
	).Scan(&item.ID, &item.UserID, &item.ShopItemID, &item.Quantity, &item.CreatedAt)
	if err != nil {
		return nil, err
	}
	// Name wird hier nicht gebraucht (nur für GetCartByUserID via JOIN)
	return item, nil
}

// AddToCart legt einen neuen Warenkorbeintrag an.
func (r *PostgresShopRepository) AddToCart(userID, shopItemID, quantity int) error {
	_, err := r.db.Exec(
		`INSERT INTO cart_items (user_id, shop_item_id, quantity) VALUES ($1, $2, $3)`,
		userID, shopItemID, quantity,
	)
	return err
}

// UpdateCartQuantity aktualisiert die Quantity eines Warenkorbeintrags.
func (r *PostgresShopRepository) UpdateCartQuantity(userID, shopItemID, newQuantity int) error {
	_, err := r.db.Exec(
		`UPDATE cart_items SET quantity = $1 WHERE user_id = $2 AND shop_item_id = $3`,
		newQuantity, userID, shopItemID,
	)
	return err
}

// RemoveCartEntry löscht einen einzelnen Warenkorbeintrag komplett.
func (r *PostgresShopRepository) RemoveCartEntry(userID, shopItemID int) error {
	_, err := r.db.Exec(
		`DELETE FROM cart_items WHERE user_id = $1 AND shop_item_id = $2`,
		userID, shopItemID,
	)
	return err
}

// ClearCart leert den gesamten Warenkorb eines Benutzers.
func (r *PostgresShopRepository) ClearCart(userID int) error {
	_, err := r.db.Exec(`DELETE FROM cart_items WHERE user_id = $1`, userID)
	return err
}
