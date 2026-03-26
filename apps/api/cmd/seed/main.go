package main

import (
	"log"

	"membership-system/api/config"
	"membership-system/api/internal/features/auth"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

	// Run migrations first
	if err := config.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Creating default users...")

	// Default users to create
	defaultUsers := []struct {
		email    string
		password string
		fullName string
		role     string
	}{
		{
			email:    "admin@system.local",
			password: "Admin123!",
			fullName: "System Administrator",
			role:     "admin",
		},
		{
			email:    "koordinator@system.local",
			password: "Koord123!",
			fullName: "Koordinator User",
			role:     "koordinator",
		},
		{
			email:    "yk1@system.local",
			password: "YK123!",
			fullName: "YK Member 1",
			role:     "yk",
		},
		{
			email:    "yk2@system.local",
			password: "YK123!",
			fullName: "YK Member 2",
			role:     "yk",
		},
		{
			email:    "yik1@system.local",
			password: "YIK123!",
			fullName: "YIK Member 1",
			role:     "yik",
		},
		{
			email:    "asil1@system.local",
			password: "Asil123!",
			fullName: "Asil Üye 1",
			role:     "asil_uye",
		},
	}

	authRepo := auth.NewRepository(db)

	for _, userData := range defaultUsers {
		// Check if user already exists
		existing, _ := authRepo.FindByEmail(nil, userData.email)
		if existing != nil {
			log.Printf("User %s already exists, skipping...", userData.email)
			continue
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v", userData.email, err)
			continue
		}

		// Create user
		user := &auth.User{
			ID:           uuid.New().String(),
			Email:        userData.email,
			PasswordHash: string(hashedPassword),
			FullName:     userData.fullName,
			Role:         auth.UserRole(userData.role),
			IsActive:     true,
		}

		if err := authRepo.Create(nil, user); err != nil {
			log.Printf("Failed to create user %s: %v", userData.email, err)
			continue
		}

		log.Printf("✓ Created user: %s (%s) - Password: %s", userData.email, userData.role, userData.password)
	}

	log.Println("Seed data created successfully!")
}
