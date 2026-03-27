package applications

import (
	"context"
	"errors"
	"strconv"

	"membership-system/api/internal/shared"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ReferenceCreator abstracts the reference service to avoid a direct import cycle.
// It is satisfied by *references.Service.
type ReferenceCreator interface {
	CreateForApplication(ctx context.Context, appID, applicantName, applicantEmail, membershipType string, refereeUserIDs []string) error
}

// Handler handles HTTP requests for the applications feature.
type Handler struct {
	service    *Service
	refCreator ReferenceCreator // optional — nil for types that don't use references
	validate   *validator.Validate
}

// NewHandler creates a new applications handler.
// refCreator may be nil if reference creation is not needed (e.g. in tests).
func NewHandler(service *Service, refCreator ReferenceCreator) *Handler {
	return &Handler{
		service:    service,
		refCreator: refCreator,
		validate:   shared.NewValidator(),
	}
}

// Submit handles POST /api/v1/applications
func (h *Handler) Submit(c *fiber.Ctx) error {
	var req CreateApplicationRequest
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

	// Get actor ID from JWT context (may be empty for public submissions)
	actorID, _ := c.Locals("userID").(string)

	result, err := h.service.Submit(c.Context(), &req, actorID)
	if err != nil {
		return mapError(c, err)
	}

	// For asil/akademik applications, create reference records and send emails.
	if h.refCreator != nil && (req.MembershipType == MembershipAsil || req.MembershipType == MembershipAkademik) {
		refereeIDs := make([]string, 0, len(req.References))
		for _, r := range req.References {
			refereeIDs = append(refereeIDs, r.UserID)
		}
		if err := h.refCreator.CreateForApplication(
			c.Context(),
			result.Application.ID,
			result.Application.ApplicantName,
			result.Application.ApplicantEmail,
			string(result.Application.MembershipType),
			refereeIDs,
		); err != nil {
			// Non-fatal: the application is created — log and return partial success.
			return shared.Created(c, fiber.Map{
				"application":      result.Application,
				"repeat_applicant": result.RepeatApplicant,
				"previous_app_id":  result.PreviousAppID,
				"warning":          "Application created but reference emails could not be sent: " + err.Error(),
			})
		}
	}

	return shared.Created(c, fiber.Map{
		"application":      result.Application,
		"repeat_applicant": result.RepeatApplicant,
		"previous_app_id":  result.PreviousAppID,
	})
}

// GetByID handles GET /api/v1/applications/:id
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Application ID is required")
	}

	requestorRole, _ := c.Locals("userRole").(string)

	app, err := h.service.GetByID(c.Context(), id, requestorRole)
	if err != nil {
		return mapError(c, err)
	}

	return shared.Success(c, app)
}

// ListAll handles GET /api/v1/applications
func (h *Handler) ListAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	filters := ApplicationFilters{
		MembershipType: c.Query("membership_type"),
		Status:         c.Query("status"),
		Search:         c.Query("search"),
		Page:           page,
		PageSize:       pageSize,
	}

	result, err := h.service.ListAll(c.Context(), filters)
	if err != nil {
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return shared.Success(c, result)
}

// GetTimeline handles GET /api/v1/applications/:id/timeline
func (h *Handler) GetTimeline(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Application ID is required")
	}

	timeline, err := h.service.GetTimeline(c.Context(), id)
	if err != nil {
		return mapError(c, err)
	}

	return shared.Success(c, timeline)
}

// GetRedHistory handles GET /api/v1/applications/:id/red-history
func (h *Handler) GetRedHistory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Application ID is required")
	}

	history, err := h.service.GetRedHistory(c.Context(), id)
	if err != nil {
		return mapError(c, err)
	}

	return shared.Success(c, history)
}

// mapError converts domain errors to HTTP responses.
func mapError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, shared.ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound):
		return shared.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Application not found")
	case errors.Is(err, shared.ErrApplicationTerminated):
		return shared.Error(c, fiber.StatusConflict, "APPLICATION_TERMINATED", err.Error())
	case errors.Is(err, shared.ErrInvalidTransition):
		return shared.Error(c, fiber.StatusUnprocessableEntity, "INVALID_TRANSITION", err.Error())
	case errors.Is(err, shared.ErrForbidden):
		return shared.Error(c, fiber.StatusForbidden, "FORBIDDEN", err.Error())
	default:
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
