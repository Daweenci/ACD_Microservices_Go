package core

import "time"

// CartItem repräsentiert einen Eintrag im Warenkorb.
// Der Service speichert nur IDs und Mengen – keine Produktdaten.
type CartItem struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ItemID    int       `json:"item_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

// CartRepository definiert die Datenbankoperationen für den Warenkorb.
type CartRepository interface {
	GetByUserID(userID int) ([]CartItem, error)
	GetEntry(userID, itemID int) (*CartItem, error)
	Add(userID, itemID, quantity int) error
	UpdateQuantity(userID, itemID, newQuantity int) error
	RemoveEntry(userID, itemID int) error
	ClearCart(userID int) error
}

// AddToCartRequest enthält die Eingabedaten für POST /cart.
type AddToCartRequest struct {
	ItemID   int `json:"item_id"`
	Quantity int `json:"quantity"`
}

// UpdateCartRequest enthält den Delta-Wert für PATCH /cart/item/{id}.
// Positiv → hinzufügen, negativ → abziehen, erreicht 0 oder weniger → löschen.
type UpdateCartRequest struct {
	Delta int `json:"delta"`
}
