package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	pass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(pass), err
}

func CheckPasswordHash(password, hashstring string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashstring), []byte(password))
}
