package core

import (
	"errors"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// ValidateJWT validiert einen JWT-Token lokal anhand des öffentlichen EC-Schlüssels.
func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	pubKeyPEM := os.Getenv("JWT_PUBLIC_KEY")
	if pubKeyPEM == "" {
		return nil, errors.New("JWT_PUBLIC_KEY nicht gesetzt")
	}

	pubKeyPEM = strings.ReplaceAll(pubKeyPEM, `\n`, "\n")

	publicKey, err := jwt.ParseECPublicKeyFromPEM([]byte(pubKeyPEM))
	if err != nil {
		return nil, errors.New("öffentlicher schlüssel ungültig: " + err.Error())
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, errors.New("ungültige signaturmethode")
		}
		return publicKey, nil
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
