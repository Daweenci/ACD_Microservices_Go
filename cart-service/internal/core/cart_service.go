package core

import "errors"

// CartService enthält die Geschäftslogik für den Warenkorb.
type CartService struct {
	repo CartRepository
}

func NewCartService(repo CartRepository) *CartService {
	return &CartService{repo: repo}
}

// GetCart gibt den Warenkorb des Benutzers zurück.
func (s *CartService) GetCart(userID int) ([]CartItem, error) {
	return s.repo.GetByUserID(userID)
}

// AddToCart fügt ein Item zum Warenkorb hinzu.
// Existiert das Item bereits, wird die Quantity erhöht.
func (s *CartService) AddToCart(userID int, req AddToCartRequest) error {
	if req.ItemID <= 0 {
		return errors.New("ungültige item_id")
	}
	if req.Quantity <= 0 {
		return errors.New("quantity muss größer als 0 sein")
	}

	existing, err := s.repo.GetEntry(userID, req.ItemID)
	if err == nil && existing != nil {
		return s.repo.UpdateQuantity(userID, req.ItemID, existing.Quantity+req.Quantity)
	}

	return s.repo.Add(userID, req.ItemID, req.Quantity)
}

// UpdateCartItem wendet einen Delta-Wert auf einen Warenkorbeintrag an.
// Positiv → hinzufügen, negativ → abziehen.
// Fällt die neue Quantity auf 0 oder darunter, wird der Eintrag gelöscht.
func (s *CartService) UpdateCartItem(userID, itemID, delta int) error {
	if delta == 0 {
		return errors.New("delta darf nicht 0 sein")
	}

	existing, err := s.repo.GetEntry(userID, itemID)
	if err != nil || existing == nil {
		if delta > 0 {
			return s.repo.Add(userID, itemID, delta)
		}
		return errors.New("item nicht im warenkorb")
	}

	newQuantity := existing.Quantity + delta
	if newQuantity <= 0 {
		return s.repo.RemoveEntry(userID, itemID)
	}
	return s.repo.UpdateQuantity(userID, itemID, newQuantity)
}

// RemoveFromCart entfernt einen Warenkorbeintrag vollständig.
func (s *CartService) RemoveFromCart(userID, itemID int) error {
	return s.repo.RemoveEntry(userID, itemID)
}

// ClearCart leert den gesamten Warenkorb des Benutzers.
func (s *CartService) ClearCart(userID int) error {
	return s.repo.ClearCart(userID)
}
