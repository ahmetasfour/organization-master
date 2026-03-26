package config

import (
	"fmt"
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

	return &cfg, nil
}
