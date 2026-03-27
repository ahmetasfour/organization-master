package shared

import "github.com/gofiber/fiber/v2"

// APIResponse represents the unified API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents error details in API responses
type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Success sends a successful API response
func Success(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

// Error sends an error API response
func Error(c *fiber.Ctx, status int, code string, message string) error {
	return c.Status(status).JSON(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// ValidationError sends a validation error response with field-specific errors.
// Returns HTTP 422 Unprocessable Entity.
func ValidationError(c *fiber.Ctx, fields map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Fields:  fields,
		},
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "BAD_REQUEST",
			Message: message,
		},
	})
}

// Created sends a 201 Created response
func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// TokenExpired sends a 410 Gone response for expired tokens
func TokenExpired(c *fiber.Ctx) error {
	return c.Status(fiber.StatusGone).JSON(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "TOKEN_EXPIRED",
			Message: "This link has expired",
		},
	})
}

// TokenUsed sends a 409 Conflict response for already-used tokens
func TokenUsed(c *fiber.Ctx) error {
	return c.Status(fiber.StatusConflict).JSON(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "TOKEN_USED",
			Message: "This link has already been used",
		},
	})
}
