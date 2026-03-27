package webpublish

import (
	"errors"

	"membership-system/api/internal/shared"

	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for web publish consent
type Handler struct {
	service *Service
}

// NewHandler creates a new web publish handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RecordConsent handles POST /api/v1/applications/:id/publish-consent
func (h *Handler) RecordConsent(c *fiber.Ctx) error {
	applicationID := c.Params("id")
	if applicationID == "" {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Application ID is required")
	}

	adminID := c.Locals("userID").(string)

	var req RecordConsentRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "Invalid request body")
	}

	response, err := h.service.RecordConsent(c.Context(), applicationID, &req, adminID)
	if err != nil {
		if errors.Is(err, ErrApplicationNotFound) {
			return shared.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Application not found")
		}
		if errors.Is(err, ErrApplicationNotAccepted) {
			return shared.Error(c, fiber.StatusUnprocessableEntity, "INVALID_STATUS", "Consent can only be recorded for accepted applications")
		}
		if errors.Is(err, ErrConsentAlreadyRecorded) {
			return shared.Error(c, fiber.StatusConflict, "ALREADY_RECORDED", "Consent has already been recorded for this application")
		}
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record consent")
	}

	return shared.Success(c, response)
}

// GetPublishedMembers handles GET /api/v1/members (public endpoint)
func (h *Handler) GetPublishedMembers(c *fiber.Ctx) error {
	members, err := h.service.GetPublishedMembers(c.Context())
	if err != nil {
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve members")
	}

	return shared.Success(c, members)
}

// GetConsentStatus handles GET /api/v1/applications/:id/publish-consent
func (h *Handler) GetConsentStatus(c *fiber.Ctx) error {
	applicationID := c.Params("id")
	if applicationID == "" {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Application ID is required")
	}

	status, err := h.service.GetConsentStatus(c.Context(), applicationID)
	if err != nil {
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve consent status")
	}

	if status == nil {
		return shared.Success(c, fiber.Map{
			"application_id": applicationID,
			"recorded":       false,
		})
	}

	return shared.Success(c, status)
}
