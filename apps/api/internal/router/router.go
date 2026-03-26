package router

import (
	"membership-system/api/internal/features/applications"
	"membership-system/api/internal/features/auth"
	"membership-system/api/internal/features/consultations"
	"membership-system/api/internal/features/logs"
	"membership-system/api/internal/features/references"
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
}
