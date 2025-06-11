package auth

import (
	"fmt"
	"os"
	"time"

	models "github.com/Faizan2005/Golang_Coding_Assessment_Makerble/models"
	jwt "github.com/golang-jwt/jwt/v5"
)

var secret = os.Getenv("JWT_SECRET")

func GenerateToken(u *models.User) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": u.ID,
		"iss": "Expense-Tracker",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	fmt.Printf("Token Claims added %+v\n", claims)

	TokenString, err := claims.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return TokenString, nil
}
