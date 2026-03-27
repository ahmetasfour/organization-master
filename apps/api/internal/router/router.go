package router

import (
	"membership-system/api/internal/features/applications"
	"membership-system/api/internal/features/auth"
	"membership-system/api/internal/features/consultations"
	"membership-system/api/internal/features/honorary"
	"membership-system/api/internal/features/logs"
	"membership-system/api/internal/features/references"
	"membership-system/api/internal/features/reputation"
	"membership-system/api/internal/features/voting"
	"membership-system/api/internal/features/webpublish"
	"membership-system/api/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	app *fiber.App,
	authHandler *auth.Handler,
	authService *auth.Service,
	logRepo *logs.Repository,
	appHandler *applications.Handler,
	refHandler *references.Handler,
	consultHandler *consultations.Handler,
	reputationHandler *reputation.Handler,
	votingHandler *voting.Handler,
	honoraryHandler *honorary.Handler,
	webpublishHandler *webpublish.Handler,
) {
	// Apply global middleware
	app.Use(middleware.CORSMiddleware())
	app.Use(middleware.AuditMiddleware(logRepo))

	// API v1 group
	api := app.Group("/api/v1")

	// Health check (no auth required)
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Auth routes (no auth middleware)
	authGroup := api.Group("/auth")
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", middleware.AuthMiddleware(authService), authHandler.Logout)

	// ─── Public token-response routes (no auth required) ───────────────────────
	refGroup := api.Group("/ref/respond")
	refGroup.Get("/:token", refHandler.GetFormData)
	refGroup.Post("/:token", refHandler.SubmitResponse)

	// Public application submission
	api.Post("/applications", appHandler.Submit)

	// Public members list (no auth required)
	api.Get("/members", webpublishHandler.GetPublishedMembers)

	// Protected routes (require authentication)
	protected := api.Group("", middleware.AuthMiddleware(authService))

	// Application routes with RBAC
	protected.Get("/applications", middleware.YKOrKoordinator(), appHandler.ListAll)
	protected.Get("/applications/:id", appHandler.GetByID)
	protected.Get("/applications/:id/timeline", middleware.YKOrAdmin(), appHandler.GetTimeline)
	protected.Get("/applications/:id/red-history", middleware.YKOrAdmin(), appHandler.GetRedHistory)

	// Reference resend — koordinator or admin only
	protected.Post("/applications/:id/references/resend/:refId",
		middleware.KoordinatorOnly(),
		refHandler.ResendToken,
	)

	// ─── Public consultation token-response routes (no auth required) ──────────
	consultGroup := api.Group("/consult/respond")
	consultGroup.Get("/:token", consultHandler.GetFormData)
	consultGroup.Post("/:token", consultHandler.SubmitResponse)

	// Consultation management — protected
	protected.Post("/applications/:id/consultations",
		middleware.KoordinatorOnly(),
		consultHandler.AddConsultees,
	)
	protected.Get("/applications/:id/consultations",
		middleware.YKOrKoordinator(),
		consultHandler.ListForApplication,
	)

	// ─── Public reputation token-response routes (no auth required) ────────────
	repGroup := api.Group("/reputation/respond")
	repGroup.Get("/:token", reputationHandler.GetFormData)
	repGroup.Post("/:token", reputationHandler.SubmitResponse)

	// Reputation management — protected
	protected.Post("/applications/:id/reputation/contacts",
		middleware.YKOrKoordinator(),
		reputationHandler.AddContacts,
	)
	protected.Get("/applications/:id/reputation",
		middleware.YKOrKoordinator(),
		reputationHandler.GetStatus,
	)

	// ─── Voting routes ──────────────────────────────────────────────────────────
	// GET votes summary (yk + admin)
	protected.Get("/applications/:id/votes",
		middleware.YKOrAdmin(),
		votingHandler.GetVotes,
	)

	// POST votes by stage
	protected.Post("/applications/:id/votes/yk-prelim",
		middleware.YKOnly(),
		votingHandler.CastVotePrelim,
	)
	protected.Post("/applications/:id/votes/yik",
		middleware.YIKOnly(),
		votingHandler.CastVoteYIK,
	)
	protected.Post("/applications/:id/votes/yk-final",
		middleware.YKOnly(),
		votingHandler.CastVoteFinal,
	)

	// ─── Honorary Membership routes ─────────────────────────────────────────────
	// POST honorary proposal (asil_uye + yik_uye only)
	protected.Post("/honorary/propose",
		honoraryHandler.Propose,
	)

	// GET all honorary proposals (yk + admin)
	protected.Get("/honorary",
		honoraryHandler.ListProposals,
	)

	// ─── Web Publish Consent routes ─────────────────────────────────────────────
	// POST/GET web publish consent (admin only)
	protected.Post("/applications/:id/publish-consent",
		middleware.AdminOnly(),
		webpublishHandler.RecordConsent,
	)
	protected.Get("/applications/:id/publish-consent",
		middleware.AdminOnly(),
		webpublishHandler.GetConsentStatus,
	)
}
