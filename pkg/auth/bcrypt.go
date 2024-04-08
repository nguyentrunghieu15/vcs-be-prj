package auth

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

type BcryptService struct{}

func (*BcryptService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (*BcryptService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Printf("BcryptService: Error check %v", err)
	}
	return err == nil
}
