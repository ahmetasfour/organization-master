package middleware

import (
	"strings"

	"membership-system/api/internal/features/auth"
	"membership-system/api/internal/shared"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(authService *auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return shared.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
		}

		// Check for Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return shared.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format")
		}

		tokenString := parts[1]

		// Validate token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			return shared.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
		}

		// Store user info in context
		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)
		c.Locals("userRole", claims.Role)

		return c.Next()
	}
}
