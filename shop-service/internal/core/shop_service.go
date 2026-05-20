package core

import "errors"

// ShopService enthält die Geschäftslogik für Shop und Warenkorb.
type ShopService struct {
	repo ShopRepository
}

func NewShopService(repo ShopRepository) *ShopService {
	return &ShopService{repo: repo}
}

// GetShopItems gibt alle Shop-Items zurück, optional gefiltert.
func (s *ShopService) GetShopItems(filter ShopFilter) ([]ShopItem, error) {
	return s.repo.GetAllItems(filter)
}

// AddToCart fügt ein Item zum Warenkorb hinzu.
// Ist das Item bereits im Warenkorb → Quantity erhöhen statt neuen Eintrag anlegen.
func (s *ShopService) AddToCart(userID int, req AddToCartRequest) error {
	if req.Quantity <= 0 {
		return errors.New("quantity muss größer als 0 sein")
	}

	// Prüfen ob das Shop-Item überhaupt existiert
	if _, err := s.repo.GetItemByID(req.ShopItemID); err != nil {
		return errors.New("shop-item nicht gefunden")
	}

	// Bereits im Warenkorb? → erhöhen
	existing, err := s.repo.GetCartEntry(userID, req.ShopItemID)
	if err == nil && existing != nil {
		return s.repo.UpdateCartQuantity(userID, req.ShopItemID, existing.Quantity+req.Quantity)
	}

	// Neu anlegen
	return s.repo.AddToCart(userID, req.ShopItemID, req.Quantity)
}

// GetCart gibt den Warenkorb des Benutzers zurück.
func (s *ShopService) GetCart(userID int) ([]CartItem, error) {
	return s.repo.GetCartByUserID(userID)
}

// RemoveFromCart zieht quantity vom Warenkorbeintrag ab.
//   - quantityProvided=false → genau 1 abziehen
//   - quantity=0             → nichts tun
//   - quantity >= vorhandene → Eintrag komplett löschen
func (s *ShopService) RemoveFromCart(userID, shopItemID int, quantity int, quantityProvided bool) error {
	existing, err := s.repo.GetCartEntry(userID, shopItemID)
	if err != nil || existing == nil {
		return errors.New("item nicht im warenkorb")
	}

	// quantity=0 explizit angegeben → nichts tun
	if quantityProvided && quantity == 0 {
		return nil
	}

	toRemove := 1
	if quantityProvided {
		toRemove = quantity
	}

	newQuantity := existing.Quantity - toRemove

	// Quantity <= 0 → Eintrag komplett entfernen
	if newQuantity <= 0 {
		return s.repo.RemoveCartEntry(userID, shopItemID)
	}

	return s.repo.UpdateCartQuantity(userID, shopItemID, newQuantity)
}

// ClearCart leert den gesamten Warenkorb des Benutzers.
func (s *ShopService) ClearCart(userID int) error {
	return s.repo.ClearCart(userID)
}

func (s *ShopService) UpdateCartItemQuantity(userID, shopItemID, delta int) error {
	if delta == 0 {
		return nil
	}

	existing, err := s.repo.GetCartEntry(userID, shopItemID)
	if err != nil || existing == nil {
		if delta > 0 {
			if _, err := s.repo.GetItemByID(shopItemID); err != nil {
				return errors.New("shop-item nicht gefunden")
			}
			return s.repo.AddToCart(userID, shopItemID, delta)
		}
		return errors.New("item nicht im warenkorb")
	}

	newQuantity := existing.Quantity + delta
	if newQuantity <= 0 {
		return s.repo.RemoveCartEntry(userID, shopItemID)
	}
	return s.repo.UpdateCartQuantity(userID, shopItemID, newQuantity)
}
