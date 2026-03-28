package references

import (
	"errors"

	"membership-system/api/internal/shared"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Handler handles HTTP requests for the references feature.
type Handler struct {
	service  *Service
	validate *validator.Validate
}

// NewHandler creates a new references handler.
func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: shared.NewValidator(),
	}
}

// GetFormData handles GET /api/v1/ref/respond/:token
// Public endpoint — no authentication required.
func (h *Handler) GetFormData(c *fiber.Ctx) error {
	rawToken := c.Params("token")
	if rawToken == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "Token gereklidir")
	}

	data, err := h.service.GetFormData(c.Context(), rawToken)
	if err != nil {
		return mapRefError(c, err)
	}

	return shared.Success(c, data)
}

// SubmitResponse handles POST /api/v1/ref/respond/:token
// Public endpoint — no authentication required.
func (h *Handler) SubmitResponse(c *fiber.Ctx) error {
	rawToken := c.Params("token")
	if rawToken == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "Token gereklidir")
	}

	var req ReferenceResponseRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "Geçersiz istek formatı")
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
		return mapRefError(c, err)
	}

	return shared.Success(c, fiber.Map{"message": "Yanıtınız kaydedildi. Katkınız için teşekkür ederiz."})
}

// ResendToken handles POST /api/v1/applications/:id/references/resend/:refId
// Requires koordinator or admin role.
func (h *Handler) ResendToken(c *fiber.Ctx) error {
	refID := c.Params("refId")
	if refID == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_REF_ID", "Referans ID'si gereklidir")
	}

	// Load the reference to get referee info
	ref, err := h.service.repo.FindByID(c.Context(), refID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Referans bulunamadı")
		}
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	// Load application for context
	type appRow struct {
		ApplicantName  string `gorm:"column:applicant_name"`
		MembershipType string `gorm:"column:membership_type"`
	}
	var app appRow
	if err := h.service.db.WithContext(c.Context()).
		Table("applications").
		Select("applicant_name", "membership_type").
		Where("id = ?", ref.ApplicationID).
		First(&app).Error; err != nil {
		return shared.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Başvuru bulunamadı")
	}

	if err := h.service.ResendToken(
		c.Context(),
		refID,
		ref.RefereeName,
		ref.RefereeEmail,
		app.ApplicantName,
		app.MembershipType,
	); err != nil {
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return shared.Success(c, fiber.Map{"message": "Referans e-postası yeniden gönderildi."})
}

// GetReplacementFormData handles GET /api/v1/ref/replace/:token
// Public endpoint — no authentication required.
func (h *Handler) GetReplacementFormData(c *fiber.Ctx) error {
	rawToken := c.Params("token")
	if rawToken == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "Token gereklidir")
	}

	data, err := h.service.GetReplacementFormData(c.Context(), rawToken)
	if err != nil {
		return mapRefError(c, err)
	}

	return shared.Success(c, data)
}

// SubmitReplacement handles POST /api/v1/ref/replace/:token
// Public endpoint — no authentication required.
func (h *Handler) SubmitReplacement(c *fiber.Ctx) error {
	rawToken := c.Params("token")
	if rawToken == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "Token gereklidir")
	}

	var req SubmitReplacementRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "Geçersiz istek formatı")
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

	if err := h.service.SubmitReplacement(c.Context(), rawToken, &req); err != nil {
		return mapRefError(c, err)
	}

	return shared.Success(c, fiber.Map{"message": "Yeni referans kaydedildi ve e-posta gönderildi."})
}

// ─── error mapper ──────────────────────────────────────────────────────────────

func mapRefError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, shared.ErrTokenExpired):
		return shared.Error(c, fiber.StatusGone, "TOKEN_EXPIRED", "Bu referans bağlantısının süresi dolmuştur.")
	case errors.Is(err, shared.ErrTokenUsed):
		return shared.Error(c, fiber.StatusConflict, "TOKEN_USED", "Bu bağlantı daha önce kullanılmıştır.")
	case errors.Is(err, shared.ErrNotFound):
		return shared.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Kaynak bulunamadı.")
	case errors.Is(err, shared.ErrApplicationTerminated):
		return shared.Error(c, fiber.StatusConflict, "APPLICATION_TERMINATED", "Bu başvuru sonuçlandırılmıştır.")
	default:
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
