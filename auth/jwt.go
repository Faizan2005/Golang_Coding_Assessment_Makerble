package auth

import (
	"fmt"
	"os"
	"time"

	models "github.com/Faizan2005/Golang_Coding_Assessment_Makerble/models"
	jwt "github.com/golang-jwt/jwt/v5"
)

// GenerateToken creates a new JWT for the given user.
// The token includes user ID, role, issuer, and expiration time.
func GenerateToken(u *models.User) (string, error) {
	// Retrieve the JWT secret from environment variables.
	// This secret is crucial for signing the token and must be kept secure.
	secret := os.Getenv("JWT_SECRET")

	// Debugging: Print the secret to ensure it's loaded correctly.
	// In a production environment, you might want to remove or secure this debug output.
	fmt.Printf("DEBUG: JWT_SECRET in GenerateToken is: '%s' (Length: %d)\n", secret, len(secret))

	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET environment variable not set")
	}

	// Define the claims to be included in the JWT.
	// These claims carry information about the user and the token itself.
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  u.ID,                                  // "sub" (subject) is a standard claim for the principal (user) of the JWT.
		"role": u.Role,                                // Custom claim to store the user's role for authorization.
		"iss":  "Hospital-Portal",                     // "iss" (issuer) identifies the principal that issued the JWT.
		"exp":  time.Now().Add(time.Hour * 24).Unix(), // "exp" (expiration time) after which the JWT must not be accepted for processing (24 hours from now).
		"iat":  time.Now().Unix(),                     // "iat" (issued at time) identifies the time at which the JWT was issued.
	})

	// For debugging: Print the claims before signing.
	fmt.Printf("Token Claims added %+v\n", claims)

	// Sign the token with the secret key.
	// The secret converts to a byte slice as required by the signing method.
	tokenString, err := claims.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
