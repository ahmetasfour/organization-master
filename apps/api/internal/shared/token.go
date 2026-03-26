package shared

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// TokenResult contains the generated token data
type TokenResult struct {
	RawToken    string
	HashedToken string
	ExpiresAt   time.Time
}

// GenerateToken generates a unique token for entity verification (references, password resets, etc.)
// Returns:
// - rawToken: UUID v4 to be sent in email URLs (NEVER stored in DB)
// - hashedToken: SHA-256 hash to be stored in DB
// - expiresAt: 48 hours from now
func GenerateToken() *TokenResult {
	// Generate UUID v4 as raw token
	rawToken := uuid.New().String()

	// Compute SHA-256 hash
	hash := sha256.Sum256([]byte(rawToken))
	hashedToken := hex.EncodeToString(hash[:])

	// Set expiry to 48 hours
	expiresAt := time.Now().Add(48 * time.Hour)

	return &TokenResult{
		RawToken:    rawToken,
		HashedToken: hashedToken,
		ExpiresAt:   expiresAt,
	}
}

// HashToken computes SHA-256 hash of a raw token for verification
func HashToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}
