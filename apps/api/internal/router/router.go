package router

import (
	"membership-system/api/config"
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
	cfg *config.Config,
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
	logsHandler *logs.Handler,
) {
	// Apply global middleware
	app.Use(middleware.SecurityHeadersMiddleware())
	app.Use(middleware.DynamicCORSMiddleware(cfg))
	app.Use(middleware.AuditMiddleware(logRepo))

	// API v1 group
	api := app.Group("/api/v1")

	// Health check (no auth required)
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Auth routes (no auth middleware, but rate limited)
	authGroup := api.Group("/auth")
	authGroup.Post("/login", middleware.LoginRateLimiter(), authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", middleware.AuthMiddleware(authService), authHandler.Logout)

	// ─── Public token-response routes (rate limited) ───────────────────────────
	refGroup := api.Group("/ref/respond")
	refGroup.Get("/:token", middleware.PublicTokenRateLimiter(), refHandler.GetFormData)
	refGroup.Post("/:token", middleware.PublicTokenRateLimiter(), refHandler.SubmitResponse)

	// Public replacement reference routes (rate limited)
	replaceGroup := api.Group("/ref/replace")
	replaceGroup.Get("/:token", middleware.PublicTokenRateLimiter(), refHandler.GetReplacementFormData)
	replaceGroup.Post("/:token", middleware.PublicTokenRateLimiter(), refHandler.SubmitReplacement)

	// Public application submission
	api.Post("/applications", appHandler.Submit)

	// Public members list (no auth required)
	api.Get("/members", webpublishHandler.GetPublishedMembers)

	// Protected routes (require authentication)
	protected := api.Group("", middleware.AuthMiddleware(authService))

	// User management routes (admin only)
	protected.Get("/users", middleware.AdminOnly(), authHandler.ListUsers)
	protected.Post("/users", middleware.AdminOnly(), authHandler.CreateUser)
	protected.Get("/users/active", authHandler.ListActiveUsers) // All authenticated users
	protected.Get("/users/:id", middleware.AdminOnly(), authHandler.GetUser)
	protected.Patch("/users/:id", middleware.AdminOnly(), authHandler.UpdateUser)

	// Application routes with RBAC
	protected.Get("/applications", middleware.YKOrKoordinator(), appHandler.ListAll)
	protected.Get("/applications/:id", appHandler.GetByID)
	protected.Get("/applications/:id/timeline", middleware.YKOrAdmin(), appHandler.GetTimeline)
	protected.Get("/applications/:id/red-history", middleware.YKOrAdmin(), appHandler.GetRedHistory)
	protected.Patch("/applications/:id/advance", middleware.YKOrKoordinator(), appHandler.Advance)

	// Reference resend — koordinator or admin only
	protected.Post("/applications/:id/references/resend/:refId",
		middleware.KoordinatorOnly(),
		refHandler.ResendToken,
	)

	// ─── Public consultation token-response routes (rate limited) ──────────────
	consultGroup := api.Group("/consult/respond")
	consultGroup.Get("/:token", middleware.PublicTokenRateLimiter(), consultHandler.GetFormData)
	consultGroup.Post("/:token", middleware.PublicTokenRateLimiter(), consultHandler.SubmitResponse)

	// Consultation management — protected
	protected.Post("/applications/:id/consultations",
		middleware.KoordinatorOnly(),
		consultHandler.AddConsultees,
	)
	protected.Get("/applications/:id/consultations",
		middleware.YKOrKoordinator(),
		consultHandler.ListForApplication,
	)

	// ─── Public reputation token-response routes (rate limited) ────────────────
	repGroup := api.Group("/reputation/respond")
	repGroup.Get("/:token", middleware.PublicTokenRateLimiter(), reputationHandler.GetFormData)
	repGroup.Post("/:token", middleware.PublicTokenRateLimiter(), reputationHandler.SubmitResponse)

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

	// ─── Logs routes ─────────────────────────────────────────────────────────────
	// GET logs list (yk, koordinator, admin only)
	protected.Get("/logs",
		middleware.YKOrKoordinator(),
		logsHandler.List,
	)
	// GET single log by ID (yk, koordinator, admin only)
	protected.Get("/logs/:id",
		middleware.YKOrKoordinator(),
		logsHandler.GetByID,
	)
}
