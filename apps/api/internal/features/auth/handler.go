package auth

import (
	"membership-system/api/internal/shared"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for authentication
type Handler struct {
	service  *Service
	validate *validator.Validate
}

// NewHandler creates a new auth handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: shared.NewValidator(),
	}
}

// Login handles the login request
// POST /api/v1/auth/login
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		fields := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			fields[err.Field()] = err.Tag()
		}
		return shared.ValidationError(c, fields)
	}

	result, err := h.service.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		if err == shared.ErrInvalidCredentials {
			return shared.Error(c, fiber.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		}
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "An error occurred during login")
	}

	return shared.Success(c, result)
}

// Refresh handles the refresh token request
// POST /api/v1/auth/refresh
func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		fields := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			fields[err.Field()] = err.Tag()
		}
		return shared.ValidationError(c, fields)
	}

	result, err := h.service.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		if err == shared.ErrUnauthorized {
			return shared.Error(c, fiber.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired refresh token")
		}
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "An error occurred during token refresh")
	}

	return shared.Success(c, result)
}

// Logout handles the logout request
// POST /api/v1/auth/logout
func (h *Handler) Logout(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("userID").(string)

	err := h.service.Logout(c.Context(), userID)
	if err != nil {
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "An error occurred during logout")
	}

	return shared.NoContent(c)
}
