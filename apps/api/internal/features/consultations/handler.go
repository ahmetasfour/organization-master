package consultations

import (
	"errors"

	"membership-system/api/internal/shared"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for the consultations feature.
type Handler struct {
	service  *Service
	validate *validator.Validate
}

// NewHandler creates a new consultations handler.
func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: shared.NewValidator(),
	}
}

// AddConsultees handles POST /api/v1/applications/:id/consultations
// Requires koordinator or admin role.
func (h *Handler) AddConsultees(c *fiber.Ctx) error {
	appID := c.Params("id")
	if appID == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_ID", "Application ID is required")
	}

	var req AddConsultationsRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "Invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		fields := make(map[string]string)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, e := range ve {
				fields[e.Field()] = e.Tag()
			}
		}
		return shared.ValidationError(c, fields)
	}

	koordinatorID, _ := c.Locals("userID").(string)

	if err := h.service.AddConsultees(c.Context(), appID, &req, koordinatorID); err != nil {
		return mapConsultError(c, err)
	}

	return shared.Success(c, fiber.Map{
		"message": "Danışma talepleri gönderildi.",
	})
}

// GetFormData handles GET /api/v1/consult/respond/:token
// Public endpoint — no authentication required.
func (h *Handler) GetFormData(c *fiber.Ctx) error {
	rawToken := c.Params("token")
	if rawToken == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "Token is required")
	}

	data, err := h.service.GetFormData(c.Context(), rawToken)
	if err != nil {
		return mapConsultError(c, err)
	}

	return shared.Success(c, data)
}

// SubmitResponse handles POST /api/v1/consult/respond/:token
// Public endpoint — no authentication required.
func (h *Handler) SubmitResponse(c *fiber.Ctx) error {
	rawToken := c.Params("token")
	if rawToken == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "Token is required")
	}

	var req ConsultationResponseRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "Invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		fields := make(map[string]string)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, e := range ve {
				fields[e.Field()] = e.Tag()
			}
		}
		return shared.ValidationError(c, fields)
	}

	ipAddress := c.IP()

	if err := h.service.SubmitResponse(c.Context(), rawToken, &req, ipAddress); err != nil {
		return mapConsultError(c, err)
	}

	return shared.Success(c, fiber.Map{
		"message": "Yanıtınız kaydedildi. Katkınız için teşekkür ederiz.",
	})
}

// ListForApplication handles GET /api/v1/applications/:id/consultations
// Requires koordinator or admin role.
func (h *Handler) ListForApplication(c *fiber.Ctx) error {
	appID := c.Params("id")
	if appID == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_ID", "Application ID is required")
	}

	summaries, err := h.service.ListForApplication(c.Context(), appID)
	if err != nil {
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return shared.Success(c, summaries)
}

// ─── error mapper ─────────────────────────────────────────────────────────────

func mapConsultError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		return shared.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Resource not found")
	case errors.Is(err, shared.ErrTokenExpired):
		return shared.Error(c, fiber.StatusGone, "TOKEN_EXPIRED", "This link has expired.")
	case errors.Is(err, shared.ErrTokenUsed):
		return shared.Error(c, fiber.StatusConflict, "TOKEN_USED", "This link has already been used.")
	case errors.Is(err, shared.ErrApplicationTerminated):
		return shared.Error(c, fiber.StatusConflict, "APPLICATION_TERMINATED", "This application has already been terminated.")
	default:
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
