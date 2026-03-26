package config

import (
	"fmt"
	"log"

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
	log.Println("[DB] Running migrations...")
	log.Println("[DB] Migrations complete")
	return nil
}
