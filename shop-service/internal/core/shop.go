// Package core enthält die reine Geschäftslogik des Shop-Service.
package core

import "time"

// ShopItem repräsentiert einen Artikel im Shop.
type ShopItem struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
}

// CartItem repräsentiert einen Eintrag im Warenkorb eines Benutzers.
// Preis wird NICHT gespeichert – er kommt immer aus shop_items via JOIN.
type CartItem struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	ShopItemID int       `json:"shop_item_id"`
	Name       string    `json:"name"` // aus JOIN mit shop_items
	Quantity   int       `json:"quantity"`
	CreatedAt  time.Time `json:"created_at"`
}

// ShopRepository definiert die Schnittstelle für Shop-Datenbankoperationen.
type ShopRepository interface {
	// Shop
	GetAllItems(filter ShopFilter) ([]ShopItem, error)
	GetItemByID(id int) (*ShopItem, error)

	// Warenkorb
	GetCartByUserID(userID int) ([]CartItem, error)
	GetCartEntry(userID, shopItemID int) (*CartItem, error)
	AddToCart(userID, shopItemID, quantity int) error
	UpdateCartQuantity(userID, shopItemID, newQuantity int) error
	RemoveCartEntry(userID, shopItemID int) error
	ClearCart(userID int) error
}

// ShopFilter enthält optionale Suchparameter für GET /shop.
type ShopFilter struct {
	Name     string
	Category string
	MinPrice float64 // 0 = kein Filter
	MaxPrice float64 // 0 = kein Filter
}

// AddToCartRequest enthält die Eingabedaten für POST /cart.
type AddToCartRequest struct {
	ShopItemID int `json:"shop_item_id"`
	Quantity   int `json:"quantity"`
}

// RemoveFromCartRequest enthält die Eingabedaten für DELETE /cart/item/{id}.
// Quantity ist optional – fehlt sie, wird 1 abgezogen.
type RemoveFromCartRequest struct {
	Quantity int `json:"quantity"` // 0 bedeutet: nicht angegeben → -1
}
