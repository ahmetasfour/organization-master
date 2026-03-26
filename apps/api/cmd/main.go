package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"membership-system/api/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := config.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Membership Management System API v1.0",
		ErrorHandler: customErrorHandler,
	})

	// Global middleware
	app.Use(recover.New())

	// Health check route
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": "1.0.0",
			"env":     cfg.AppEnv,
		})
	})

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%d", cfg.AppPort)
		log.Printf("[APP] Starting server on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Printf("[APP] Server error: %v", err)
		}
	}()

	<-quit
	log.Println("[APP] Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Printf("[APP] Error during shutdown: %v", err)
	}
	log.Println("[APP] Server stopped")
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    fmt.Sprintf("HTTP_%d", code),
			"message": message,
		},
	})
}
