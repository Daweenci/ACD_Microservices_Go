package core

import "errors"

// OrderService enthält die Geschäftslogik für Bestellungen.
type OrderService struct {
	repo OrderRepository
}

// NewOrderService erstellt einen neuen OrderService.
func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// CreateOrder legt eine neue Bestellung an.
func (s *OrderService) CreateOrder(userID int, req CreateOrderRequest) (*Order, error) {
	if req.Item == "" {
		return nil, errors.New("item darf nicht leer sein")
	}
	if req.Quantity <= 0 {
		return nil, errors.New("quantity muss größer als 0 sein")
	}

	order := &Order{
		UserID:   userID,
		Item:     req.Item,
		Quantity: req.Quantity,
	}
	if err := s.repo.Create(order); err != nil {
		return nil, err
	}
	return order, nil
}

// GetOrdersByUser gibt alle Bestellungen eines Benutzers zurück.
func (s *OrderService) GetOrdersByUser(userID int) ([]Order, error) {
	return s.repo.FindByUserID(userID)
}
