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

// ─── User Management Handlers ──────────────────────────────────────────────────

// ListUsers handles GET /api/v1/users (admin only)
func (h *Handler) ListUsers(c *fiber.Ctx) error {
	var filters UserFilters
	if err := c.QueryParser(&filters); err != nil {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_QUERY", "Invalid query parameters")
	}

	result, err := h.service.ListUsers(c.Context(), filters)
	if err != nil {
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return shared.Success(c, result)
}

// CreateUser handles POST /api/v1/users (admin only)
func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "Geçersiz istek formatı")
	}

	if err := h.validate.Struct(&req); err != nil {
		fields := make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			fields[e.Field()] = e.Tag()
		}
		return shared.ValidationError(c, fields)
	}

	actorID, _ := c.Locals("userID").(string)
	user, err := h.service.CreateUser(c.Context(), &req, actorID)
	if err != nil {
		if err.Error() == "email already in use" {
			return shared.Error(c, fiber.StatusConflict, "EMAIL_EXISTS", "Bu e-posta adresi zaten kullanımda")
		}
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return shared.Created(c, fiber.Map{
		"id":        user.ID,
		"full_name": user.FullName,
		"email":     user.Email,
		"role":      user.Role,
	})
}

// GetUser handles GET /api/v1/users/:id (admin only)
func (h *Handler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Kullanıcı ID'si gereklidir")
	}

	user, err := h.service.GetUser(c.Context(), id)
	if err != nil {
		if err == shared.ErrNotFound {
			return shared.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Kullanıcı bulunamadı")
		}
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return shared.Success(c, user)
}

// UpdateUser handles PATCH /api/v1/users/:id (admin only)
func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Kullanıcı ID'si gereklidir")
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "Geçersiz istek formatı")
	}

	if err := h.validate.Struct(&req); err != nil {
		fields := make(map[string]string)
		for _, e := range err.(validator.ValidationErrors) {
			fields[e.Field()] = e.Tag()
		}
		return shared.ValidationError(c, fields)
	}

	actorID, _ := c.Locals("userID").(string)
	if err := h.service.UpdateUser(c.Context(), id, &req, actorID); err != nil {
		if err == shared.ErrNotFound {
			return shared.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Kullanıcı bulunamadı")
		}
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return shared.Success(c, fiber.Map{
		"message": "Kullanıcı başarıyla güncellendi",
	})
}

// ListActiveUsers handles GET /api/v1/users/active (for member selection in consultations)
func (h *Handler) ListActiveUsers(c *fiber.Ctx) error {
	role := c.Query("role", "")

	users, err := h.service.ListActiveByRole(c.Context(), role)
	if err != nil {
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return shared.Success(c, fiber.Map{
		"data": users,
	})
}
