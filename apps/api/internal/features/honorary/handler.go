package honorary

import (
	"errors"
	"net/http"

	"membership-system/api/internal/shared"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	service  *Service
	validate *validator.Validate
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: shared.NewValidator(),
	}
}

func (h *Handler) Propose(c *fiber.Ctx) error {
	// Get user from JWT claims
	userClaims := c.Locals("user").(*jwt.Token)
	claims := userClaims.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)

	var req ProposeRequest
	if err := c.BodyParser(&req); err != nil {
		return shared.Error(c, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
	}

	// Validate request
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

	// Create proposal
	application, err := h.service.Propose(c.Context(), req, userID)
	if err != nil {
		return shared.Error(c, http.StatusBadRequest, "PROPOSAL_FAILED", err.Error())
	}

	return shared.Created(c, map[string]interface{}{
		"application_id": application.ID,
		"message":        "Honorary proposal created successfully",
	})
}

func (h *Handler) ListProposals(c *fiber.Ctx) error {
	proposals, err := h.service.ListProposals(c.Context())
	if err != nil {
		return shared.Error(c, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch proposals")
	}

	return shared.Success(c, proposals)
}

// ProposerOnlyMiddleware ensures only asil_uye and yik_uye can access
func ProposerOnlyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userClaims := c.Locals("user").(*jwt.Token)
		claims := userClaims.Claims.(jwt.MapClaims)
		role := claims["role"].(string)

		if role != "asil_uye" && role != "yik_uye" {
			return shared.Error(c, http.StatusForbidden, "ACCESS_DENIED", "Only asil_uye and yik_uye can propose honorary members")
		}

		return c.Next()
	}
}

// YKOrAdminMiddleware ensures only YK members and admins can access
func YKOrAdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userClaims := c.Locals("user").(*jwt.Token)
		claims := userClaims.Claims.(jwt.MapClaims)
		role := claims["role"].(string)

		if role != "yk" && role != "admin" {
			return shared.Error(c, http.StatusForbidden, "ACCESS_DENIED", "Access denied")
		}

		return c.Next()
	}
}
