package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword takes a plain password and returns a bcrypt hash
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("could not hash password: %w", err)
	}
	return string(hashed), nil
}

// CheckPassword compares a plain password with its hashed version
func CheckPassword(password, hashed string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	if err != nil {
		return fmt.Errorf("password does not match")
	}
	return nil
}
