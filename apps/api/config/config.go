package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	DBHost string `mapstructure:"DB_HOST"`
	DBPort int    `mapstructure:"DB_PORT"`
	DBName string `mapstructure:"DB_NAME"`
	DBUser string `mapstructure:"DB_USER"`
	DBPass string `mapstructure:"DB_PASS"`

	JWTSecret        string `mapstructure:"JWT_SECRET"`
	JWTRefreshSecret string `mapstructure:"JWT_REFRESH_SECRET"`
	JWTAccessTTL     string `mapstructure:"JWT_ACCESS_TTL"`
	JWTRefreshTTL    string `mapstructure:"JWT_REFRESH_TTL"`

	AppBaseURL string `mapstructure:"APP_BASE_URL"`
	AppPort    int    `mapstructure:"APP_PORT"`
	AppEnv     string `mapstructure:"APP_ENV"`

	MailHost     string `mapstructure:"MAIL_HOST"`
	MailPort     int    `mapstructure:"MAIL_PORT"`
	MailFrom     string `mapstructure:"MAIL_FROM"`
	MailFromName string `mapstructure:"MAIL_FROM_NAME"`
}

// IsProduction returns true if running in production environment.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// IsDevelopment returns true if running in development environment.
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development" || c.AppEnv == ""
}

// Load reads configuration from environment and validates required fields.
func Load() (*Config, error) {
	v := viper.New()

	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", 3306)
	v.SetDefault("DB_NAME", "membership_db")
	v.SetDefault("DB_USER", "root")
	v.SetDefault("DB_PASS", "secret")
	v.SetDefault("JWT_ACCESS_TTL", "15m")
	v.SetDefault("JWT_REFRESH_TTL", "168h")
	v.SetDefault("APP_BASE_URL", "http://localhost:3000")
	v.SetDefault("APP_PORT", 8080)
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("MAIL_HOST", "localhost")
	v.SetDefault("MAIL_PORT", 1025)
	v.SetDefault("MAIL_FROM", "noreply@membership.local")
	v.SetDefault("MAIL_FROM_NAME", "Membership System")

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig()

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required environment variables
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	// Log startup environment
	log.Printf("[CONFIG] Environment: %s", cfg.AppEnv)
	if cfg.IsProduction() {
		log.Println("[CONFIG] Running in PRODUCTION mode")
	} else {
		log.Println("[CONFIG] Running in DEVELOPMENT mode")
	}

	return &cfg, nil
}

// validateConfig checks that all required configuration is present.
func validateConfig(cfg *Config) error {
	var missing []string

	// JWT secrets are always required
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if cfg.JWTRefreshSecret == "" {
		missing = append(missing, "JWT_REFRESH_SECRET")
	}

	// In production, enforce stricter validation
	if cfg.IsProduction() {
		if cfg.DBPass == "secret" || cfg.DBPass == "" {
			missing = append(missing, "DB_PASS (must be set in production)")
		}
		if !strings.HasPrefix(cfg.AppBaseURL, "https://") {
			log.Println("[WARN] APP_BASE_URL should use HTTPS in production")
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}
