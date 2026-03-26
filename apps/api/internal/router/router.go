package router

import (
	"membership-system/api/internal/features/auth"
	"membership-system/api/internal/features/logs"
	"membership-system/api/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, authHandler *auth.Handler, authService *auth.Service, logRepo *logs.Repository) {
	// Apply global middleware
	app.Use(middleware.CORSMiddleware())
	app.Use(middleware.AuditMiddleware(logRepo))

	// API v1 group
	api := app.Group("/api/v1")

	// Health check (no auth required)
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Auth routes (no auth middleware)
	authGroup := api.Group("/auth")
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", middleware.AuthMiddleware(authService), authHandler.Logout)

	// Protected routes (require authentication)
	protected := api.Group("", middleware.AuthMiddleware(authService))

	// Placeholder for future routes
	_ = protected
}
