package core

import (
	"errors"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type AuthService struct {
	repo UserRepository
}

func NewAuthService(repo UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Register erstellt einen neuen Benutzer.
func (s *AuthService) Register(req RegisterRequest) (*User, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, errors.New("alle Felder müssen ausgefüllt sein")
	}
	if !emailRegex.MatchString(req.Email) {
		return nil, errors.New("ungültiges email-format")
	}
	if len(req.Password) < 6 {
		return nil, errors.New("passwort muss mindestens 6 zeichen lang sein")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hash),
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login akzeptiert Email ODER Username als Identifier.
func (s *AuthService) Login(req LoginRequest) (string, error) {
	if req.Identifier == "" || req.Password == "" {
		return "", errors.New("identifier und passwort erforderlich")
	}

	var user *User
	var err error

	if strings.Contains(req.Identifier, "@") {
		user, err = s.repo.FindByEmail(req.Identifier)
	} else {
		user, err = s.repo.FindByUsername(req.Identifier)
	}
	if err != nil {
		return "", errors.New("benutzer nicht gefunden")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("falsches passwort")
	}

	return generateJWT(user)
}

// DeleteUser löscht den Benutzer mit der gegebenen ID.
func (s *AuthService) DeleteUser(userID int) error {
	return s.repo.Delete(userID)
}

// ChangeUsername ändert den Benutzernamen des Benutzers mit der gegebenen ID.
func (s *AuthService) ChangeUsername(userID int, req ChangeUsernameRequest) error {
	if req.NewUsername == "" {
		return errors.New("neuer benutzername darf nicht leer sein")
	}
	return s.repo.UpdateUsername(userID, req.NewUsername)
}
