package jwt

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(Userid string, UserRole string) (string, error) {
	claims := jwt.MapClaims{
		"userId": Userid,
		"role":   UserRole,
		"exp":    time.Now().Add(time.Hour * 24).Unix(), //Token expires after 1 hour
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
