package voting

import (
	"errors"

	"membership-system/api/internal/shared"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for the voting system.
type Handler struct {
	service  *Service
	validate *validator.Validate
}

// NewHandler creates a new voting handler.
func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: shared.NewValidator(),
	}
}

// ─── CastVotePrelim ───────────────────────────────────────────────────────────

// CastVotePrelim handles POST /api/v1/applications/:id/votes/yk-prelim
// Requires YK role. Casts a vote at the yk_prelim stage.
func (h *Handler) CastVotePrelim(c *fiber.Ctx) error {
	return h.castVote(c, VoteStageYKPrelim)
}

// ─── CastVoteYIK ──────────────────────────────────────────────────────────────

// CastVoteYIK handles POST /api/v1/applications/:id/votes/yik
// Requires YIK role. Casts a vote at the yik stage (Onursal only).
func (h *Handler) CastVoteYIK(c *fiber.Ctx) error {
	return h.castVote(c, VoteStageYIK)
}

// ─── CastVoteFinal ────────────────────────────────────────────────────────────

// CastVoteFinal handles POST /api/v1/applications/:id/votes/yk-final
// Requires YK role. Casts a vote at the yk_final stage.
func (h *Handler) CastVoteFinal(c *fiber.Ctx) error {
	return h.castVote(c, VoteStageYKFinal)
}

// ─── GetVotes ─────────────────────────────────────────────────────────────────

// GetVotes handles GET /api/v1/applications/:id/votes?stage=<stage>
// Requires YK or admin role. Returns vote summary for the given stage.
func (h *Handler) GetVotes(c *fiber.Ctx) error {
	appID := c.Params("id")
	if appID == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_ID", "Application ID is required")
	}

	stageParam := c.Query("stage")
	if stageParam == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_STAGE", "Query parameter 'stage' is required")
	}

	stage := VoteStage(stageParam)
	if stage != VoteStageYKPrelim && stage != VoteStageYIK && stage != VoteStageYKFinal {
		return shared.Error(c, fiber.StatusBadRequest, "INVALID_STAGE",
			"Stage must be one of: yk_prelim, yik, yk_final")
	}

	requestorRole, _ := c.Locals("userRole").(string)

	summary, err := h.service.GetVotes(c.Context(), appID, stage, requestorRole)
	if err != nil {
		return mapVoteError(c, err)
	}

	return shared.Success(c, summary)
}

// ─── shared cast helper ───────────────────────────────────────────────────────

// castVote is the shared implementation for all three CastVote* handlers.
func (h *Handler) castVote(c *fiber.Ctx, stage VoteStage) error {
	appID := c.Params("id")
	if appID == "" {
		return shared.Error(c, fiber.StatusBadRequest, "MISSING_ID", "Application ID is required")
	}

	var req CastVoteRequest
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

	// Additional semantic validation: reject requires a reason
	if req.VoteType == string(VoteReject) && len([]rune(req.Reason)) < 20 {
		return shared.ValidationError(c, map[string]string{
			"reason": "min=20 required when vote_type is reject",
		})
	}

	voterID, _ := c.Locals("userID").(string)
	voterRole, _ := c.Locals("userRole").(string)

	if err := h.service.CastVote(c.Context(), appID, voterID, voterRole, stage, &req); err != nil {
		return mapVoteError(c, err)
	}

	return shared.Success(c, fiber.Map{
		"message": "Oyunuz başarıyla kaydedildi.",
	})
}

// ─── error mapper ─────────────────────────────────────────────────────────────

func mapVoteError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		return shared.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Application not found")
	case errors.Is(err, shared.ErrDuplicateVote):
		return shared.Error(c, fiber.StatusConflict, "DUPLICATE_VOTE", "You have already voted in this stage")
	case errors.Is(err, shared.ErrForbidden):
		return shared.Error(c, fiber.StatusForbidden, "FORBIDDEN", err.Error())
	case errors.Is(err, shared.ErrApplicationTerminated):
		return shared.Error(c, fiber.StatusConflict, "APPLICATION_TERMINATED",
			"This application has already been terminated and cannot receive further votes")
	default:
		return shared.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
