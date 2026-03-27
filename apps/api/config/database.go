package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"membership-system/api/internal/features/applications"
	"membership-system/api/internal/features/auth"
	"membership-system/api/internal/features/consultations"
	"membership-system/api/internal/features/logs"
	"membership-system/api/internal/features/references"
	"membership-system/api/internal/features/reputation"
	"membership-system/api/internal/features/voting"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB(cfg *Config) (*gorm.DB, error) {
	if cfg.DBHost == "skip" {
		log.Println("[DB] Skipping database connection")
		return nil, nil
	}

	dsn := buildDSN(cfg)

	logLevel := logger.Silent
	if cfg.AppEnv == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("[DB] Connected to MySQL successfully")

	DB = db
	return db, nil
}

func buildDSN(cfg *Config) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)
}

func RunMigrations(db *gorm.DB) error {
	if db == nil {
		log.Println("[DB] Skipping migrations (no database connection)")
		return nil
	}

	log.Println("[DB] Running SQL migrations...")

	// Execute SQL migration files in order
	migrationFiles := []string{
		"migrations/001_users.sql",
		"migrations/002_applications.sql",
		"migrations/003_references.sql",
		"migrations/004_reference_responses.sql",
		"migrations/005_consultations.sql",
		"migrations/006_reputation_contacts.sql",
		"migrations/007_votes.sql",
		"migrations/008_web_publish_consents.sql",
		"migrations/009_logs.sql",
	}

	for _, migrationFile := range migrationFiles {
		if err := executeSQLFile(db, migrationFile); err != nil {
			return fmt.Errorf("failed to execute %s: %w", migrationFile, err)
		}
		log.Printf("[DB] ✓ Executed %s", filepath.Base(migrationFile))
	}

	if err := ensureWebPublishConsentColumns(db); err != nil {
		return fmt.Errorf("failed to ensure web_publish_consents schema: %w", err)
	}

	// Create trigger separately
	if err := createRejectionReasonTrigger(db); err != nil {
		log.Printf("[DB] ⚠ Failed to create trigger (may already exist): %v", err)
	} else {
		log.Println("[DB] ✓ Created rejection_reason immutability trigger")
	}

	log.Println("[DB] Running GORM AutoMigrate...")

	// Run GORM AutoMigrate for all models
	if err := db.AutoMigrate(
		&auth.User{},
		&applications.Application{},
		&references.Reference{},
		&references.ReferenceResponse{},
		&consultations.Consultation{},
		&reputation.ReputationContact{},
		&voting.Vote{},
		&logs.Log{},
	); err != nil {
		return fmt.Errorf("failed to auto-migrate models: %w", err)
	}

	log.Println("[DB] ✓ Migrations complete")
	return nil
}

// createRejectionReasonTrigger creates the trigger to enforce immutability
func createRejectionReasonTrigger(db *gorm.DB) error {
	// Drop trigger if exists
	_ = db.Exec("DROP TRIGGER IF EXISTS prevent_rejection_reason_update").Error

	// Create the trigger
	triggerSQL := `
CREATE TRIGGER prevent_rejection_reason_update
BEFORE UPDATE ON applications
FOR EACH ROW
BEGIN
    IF OLD.rejection_reason IS NOT NULL 
       AND NEW.rejection_reason IS NOT NULL 
       AND OLD.rejection_reason != NEW.rejection_reason THEN
        SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = 'rejection_reason is immutable once set';
    END IF;
END`

	return db.Exec(triggerSQL).Error
}

// executeSQLFile reads and executes a SQL file
func executeSQLFile(db *gorm.DB, filePath string) error {
	// Try multiple possible paths
	possiblePaths := []string{
		filePath,
		filepath.Join("apps/api", filePath),
		filepath.Join("../../", filePath),
	}

	var sqlBytes []byte
	var err error
	var foundPath string

	for _, path := range possiblePaths {
		sqlBytes, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}

	if err != nil {
		log.Printf("[DB] ⚠ Migration file not found: %s (skipping)", filePath)
		return nil
	}

	sql := string(sqlBytes)
	if sql == "" {
		return nil
	}

	// Execute the SQL
	if err := db.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to execute SQL from %s: %w", foundPath, err)
	}

	return nil
}

func ensureWebPublishConsentColumns(db *gorm.DB) error {
	var columnCount int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
		  AND table_name = 'web_publish_consents'
		  AND column_name = 'recorded_by'
	`).Scan(&columnCount).Error; err != nil {
		return err
	}

	if columnCount == 0 {
		if err := db.Exec(`
			ALTER TABLE web_publish_consents
			ADD COLUMN recorded_by CHAR(36) NULL
		`).Error; err != nil {
			return err
		}
	}

	var fkCount int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.table_constraints
		WHERE table_schema = DATABASE()
		  AND table_name = 'web_publish_consents'
		  AND constraint_type = 'FOREIGN KEY'
		  AND constraint_name = 'fk_web_publish_consents_recorded_by'
	`).Scan(&fkCount).Error; err != nil {
		return err
	}

	if fkCount == 0 {
		if err := db.Exec(`
			ALTER TABLE web_publish_consents
			ADD CONSTRAINT fk_web_publish_consents_recorded_by
			FOREIGN KEY (recorded_by) REFERENCES users(id)
			ON DELETE SET NULL
		`).Error; err != nil {
			return err
		}
	}

	return nil
}
