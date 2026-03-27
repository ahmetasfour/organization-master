package middleware

import (
	"membership-system/api/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSMiddleware configures CORS for the API.
// In production mode, it enforces HTTPS-only origins from APP_BASE_URL.
// In development mode, it allows localhost origins.
func CORSMiddleware() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:3001", // Overridden by DynamicCORSMiddleware
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	})
}

// DynamicCORSMiddleware creates CORS middleware based on environment configuration.
// In production, only allows APP_BASE_URL origin with HTTPS enforcement.
func DynamicCORSMiddleware(cfg *config.Config) fiber.Handler {
	var allowedOrigins string

	if cfg.IsProduction() {
		// In production, only allow the configured base URL
		allowedOrigins = cfg.AppBaseURL
	} else {
		// In development, allow localhost origins
		allowedOrigins = "http://localhost:3000,http://localhost:3001"
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	})
}
