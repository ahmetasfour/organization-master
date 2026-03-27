package middleware

import (
	"membership-system/api/internal/shared"

	"github.com/gofiber/fiber/v2"
)

// RequireRole checks if the user has one of the required roles
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("userRole")
		if userRole == nil {
			return shared.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Kimlik doğrulama gereklidir")
		}

		role := userRole.(string)
		for _, allowedRole := range roles {
			if role == allowedRole {
				return c.Next()
			}
		}

		return shared.Error(c, fiber.StatusForbidden, "FORBIDDEN", "Bu işlemi gerçekleştirmek için yetkiniz yok")
	}
}

// Named middleware shortcuts for common role combinations

// AdminOnly requires admin role
func AdminOnly() fiber.Handler {
	return RequireRole("admin")
}

// YKOnly requires yk role
func YKOnly() fiber.Handler {
	return RequireRole("yk")
}

// YIKOnly requires yik role
func YIKOnly() fiber.Handler {
	return RequireRole("yik")
}

// KoordinatorOnly requires koordinator role
func KoordinatorOnly() fiber.Handler {
	return RequireRole("koordinator")
}

// YKOrAdmin requires yk or admin role
func YKOrAdmin() fiber.Handler {
	return RequireRole("yk", "admin")
}

// YKOrKoordinator requires yk, koordinator, or admin role
func YKOrKoordinator() fiber.Handler {
	return RequireRole("yk", "koordinator", "admin")
}

// ProposerOnly requires asil_uye or yik_uye role (users who can propose applications)
func ProposerOnly() fiber.Handler {
	return RequireRole("asil_uye", "yik_uye")
}

// YKOrYIK requires yk or yik role
func YKOrYIK() fiber.Handler {
	return RequireRole("yk", "yik")
}
