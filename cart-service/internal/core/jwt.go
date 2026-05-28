package core

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// ValidateJWT validiert einen JWT-Token lokal anhand des gemeinsamen Secrets.
func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	publicKey := os.Getenv("JWT_PUBLIC_KEY")
	if publicKey == "" {
		publicKey = "dev-public-key"
	}
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("ungültige signaturmethode")
		}
		return []byte(publicKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("ungültiger token")
	}
	return claims, nil
}
