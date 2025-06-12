package auth

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware is a Fiber middleware that authenticates requests using JWTs.
// It expects a "Bearer <token>" in the Authorization header.
func JWTMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing authentication token"})
	}

	// Split the "Bearer" prefix from the actual token string.
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token format. Expected 'Bearer <token>'"})
	}

	tokenString := tokenParts[1]

	// Parse and validate the JWT.
	token, err := ParseJWT(tokenString)
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	// Extract claims from the token.
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to extract token claims"})
	}

	// Get userID from claims.
	userID, ok := claims["sub"].(string) // "sub" is a standard JWT claim for subject/user ID.
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User ID claim missing or invalid in token"})
	}

	// Get userRole from claims.
	userRole, ok := claims["role"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User role claim missing or invalid in token"})
	}

	// Store userID and userRole in Fiber's locals for subsequent handlers.
	c.Locals("userID", userID)
	c.Locals("userRole", userRole)

	return c.Next() // Continue to the next middleware or route handler.
}

// ParseJWT parses and validates a JWT string.
// It uses the JWT_SECRET environment variable for signing key verification.
func ParseJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token's signing method is HMAC.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return the secret key for validation.
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			return nil, fmt.Errorf("JWT_SECRET environment variable not set")
		}
		return []byte(jwtSecret), nil
	})

	return token, err
}

// RoleMiddleware creates a Fiber middleware that restricts access based on user roles.
// It expects userRole to be set in c.Locals by a preceding middleware (e.g., JWTMiddleware).
func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("userRole")

		if userRole == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied: User role not found in context"})
		}

		roleStr, ok := userRole.(string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Access denied: Invalid role format in context"})
		}

		// Check if the user's role is in the list of allowed roles.
		for _, allowed := range allowedRoles {
			if roleStr == allowed {
				return c.Next() // User has an allowed role, continue.
			}
		}

		// If no allowed role matches, deny access.
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied: Insufficient permissions for this action"})
	}
}
