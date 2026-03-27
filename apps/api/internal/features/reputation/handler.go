package reputation

import (
	"errors"

	"membership-system/api/internal/shared"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for the reputation screening feature.
type Handler struct {
	service  *Service
	validate *validator.Validate
}

// NewHandler creates a new reputation handler.
func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: shared.NewValidator(),
	}
}

// AddContacts handles POST /api/v1/applications/:id/reputation/contacts
// Requires yk or koordinator role.
func (h *Handler) AddContacts(c *fiber.Ctx) error {
	appID := c.Params("id")
	if appID == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_ID", "Application ID is required")
	}

	var req AddContactsRequest
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

	// Service-layer check enforces exactly 10 regardless of validator tag
	if len(req.Contacts) != 10 {
		return shared.Error(c, fiber.StatusUnprocessableEntity, "INVALID_CONTACTS_COUNT",
			"Exactly 10 contacts are required")
	}

	actorID, _ := c.Locals("userID").(string)
	actorRole, _ := c.Locals("userRole").(string)

	if err := h.service.AddContacts(c.Context(), appID, &req, actorID, actorRole); err != nil {
		return mapRepError(c, err)
	}

	return shared.Success(c, fiber.Map{
		"message": "İtibar tarama talepleri gönderildi.",
	})
}

// GetStatus handles GET /api/v1/applications/:id/reputation
// Requires yk, koordinator, or admin role.
func (h *Handler) GetStatus(c *fiber.Ctx) error {
	appID := c.Params("id")
	if appID == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_ID", "Application ID is required")
	}

	status, err := h.service.GetStatus(c.Context(), appID)
	if err != nil {
		return mapRepError(c, err)
	}

	return shared.Success(c, status)
}

// GetFormData handles GET /api/v1/reputation/respond/:token
// Public endpoint — no authentication required.
func (h *Handler) GetFormData(c *fiber.Ctx) error {
	rawToken := c.Params("token")
	if rawToken == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "Token is required")
	}

	data, err := h.service.GetFormData(c.Context(), rawToken)
	if err != nil {
		return mapRepError(c, err)
	}

	return shared.Success(c, data)
}

// SubmitResponse handles POST /api/v1/reputation/respond/:token
// Public endpoint — no authentication required.
func (h *Handler) SubmitResponse(c *fiber.Ctx) error {
	rawToken := c.Params("token")
	if rawToken == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "Token is required")
	}

	var req ContactResponseRequest
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
		return mapRepError(c, err)
	}

	return shared.Success(c, fiber.Map{
		"message": "Yanıtınız kaydedildi. Katkınız için teşekkür ederiz.",
	})
}

// ─── error mapper ─────────────────────────────────────────────────────────────

func mapRepError(c *fiber.Ctx, err error) error {
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
