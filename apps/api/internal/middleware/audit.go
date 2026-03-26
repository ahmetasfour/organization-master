package middleware

import (
	"context"
	"encoding/json"
	"time"

	"membership-system/api/internal/features/logs"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
)

// AuditMiddleware logs all non-GET requests for audit trail
func AuditMiddleware(logRepo *logs.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip GET requests
		if c.Method() == fiber.MethodGet {
			return c.Next()
		}

		// Extract data that we need for logging BEFORE starting goroutine
		// (fiber context is not safe to use in goroutines)
		userID, _ := c.Locals("userID").(string)
		userRole, _ := c.Locals("userRole").(string)
		method := c.Method()
		path := c.Path()
		ip := c.IP()
		userAgent := c.Get("User-Agent")

		// Continue with request processing
		err := c.Next()

		// Get status code after response
		statusCode := c.Response().StatusCode()

		// Log after response (fire and forget)
		go func() {
			// Determine action based on method and path
			action := "http." + method + "." + path

			// Create metadata
			metadata := map[string]interface{}{
				"method":      method,
				"path":        path,
				"status_code": statusCode,
				"ip":          ip,
				"user_agent":  userAgent,
			}

			// Convert metadata to JSON
			metadataJSON, _ := json.Marshal(metadata)

			// Prepare actor ID pointer (can be nil if not authenticated)
			var actorIDPtr *string
			if userID != "" {
				actorIDPtr = &userID
			}

			// Create log entry
			logEntry := &logs.Log{
				EntityType: "http",
				EntityID:   "", // No specific entity for generic HTTP requests
				Action:     action,
				ActorID:    actorIDPtr,
				ActorRole:  userRole,
				Metadata:   datatypes.JSON(metadataJSON),
				CreatedAt:  time.Now(),
			}

			// Write to database using background context (goroutine-safe)
			_ = logRepo.Create(context.Background(), logEntry)
		}()

		return err
	}
}
