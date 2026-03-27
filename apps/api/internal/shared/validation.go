package shared

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ─── Custom Validators ─────────────────────────────────────────────────────────

var (
	// linkedInPattern validates LinkedIn profile URLs.
	linkedInPattern = regexp.MustCompile(`^https://(www\.)?linkedin\.com/in/[a-zA-Z0-9\-]+/?$`)

	// tokenPattern validates UUID v4 format tokens.
	tokenPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

// ValidateLinkedInURL validates that a URL is a valid LinkedIn profile URL.
// Accepts: https://linkedin.com/in/username or https://www.linkedin.com/in/username
func ValidateLinkedInURL(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	if url == "" {
		return true // Empty is valid (use 'required' tag for required fields)
	}
	return linkedInPattern.MatchString(url)
}

// ValidateTokenFormat validates that a string is a valid UUID v4 token.
func ValidateTokenFormat(fl validator.FieldLevel) bool {
	token := fl.Field().String()
	if token == "" {
		return false
	}
	return tokenPattern.MatchString(strings.ToLower(token))
}

// ValidatePhotoURL validates that a URL is a valid photo URL (http/https).
func ValidatePhotoURL(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	if url == "" {
		return true // Empty is valid
	}
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// ─── Validator Instance ────────────────────────────────────────────────────────

// NewValidator creates a validator instance with custom validation rules registered.
func NewValidator() *validator.Validate {
	v := validator.New()

	// Register custom validators
	_ = v.RegisterValidation("linkedin_url", ValidateLinkedInURL)
	_ = v.RegisterValidation("token_format", ValidateTokenFormat)
	_ = v.RegisterValidation("photo_url", ValidatePhotoURL)

	return v
}

// ─── Validation Helper Functions ───────────────────────────────────────────────

// IsValidLinkedInURL checks if a URL is a valid LinkedIn profile URL.
func IsValidLinkedInURL(url string) bool {
	if url == "" {
		return true
	}
	return linkedInPattern.MatchString(url)
}

// IsValidTokenFormat checks if a string is a valid UUID v4 token.
func IsValidTokenFormat(token string) bool {
	if token == "" || len(token) < 36 {
		return false
	}
	return tokenPattern.MatchString(strings.ToLower(token))
}

// IsValidPhotoURL checks if a URL is a valid photo URL.
func IsValidPhotoURL(url string) bool {
	if url == "" {
		return true
	}
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}
