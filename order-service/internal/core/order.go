package core

import "time"

// Order repräsentiert eine Bestellung.
type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Item      string    `json:"item"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

// OrderRepository definiert die Schnittstelle zur Datenpersistenz.
type OrderRepository interface {
	Create(o *Order) error
	FindByUserID(userID int) ([]Order, error)
}

// CreateOrderRequest enthält die Eingabedaten für eine neue Bestellung.
type CreateOrderRequest struct {
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
}
