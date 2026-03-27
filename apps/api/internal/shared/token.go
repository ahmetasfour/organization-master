package shared

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

// tokenRow is an internal struct used by ValidateAndConsumeToken for generic table lookups.
type tokenRow struct {
	ID             string     `gorm:"column:id"`
	TokenExpiresAt time.Time  `gorm:"column:token_expires_at"`
	TokenUsedAt    *time.Time `gorm:"column:token_used_at"`
}

// ValidateAndConsumeToken validates a hashed token in any token-bearing table and
// atomically marks it as used. It must be called inside an existing DB transaction
// so the consume step is rolled back if the surrounding operation fails.
//
// table must be one of: "references", "consultations", "reputation_contacts".
//
// Returns:
//   - ErrNotFound  if no record matches the hash
//   - ErrTokenExpired if token_expires_at < now   (caller should respond 410)
//   - ErrTokenUsed    if token_used_at IS NOT NULL (caller should respond 409)
//   - nil on success (token marked used)
func ValidateAndConsumeToken(
	ctx context.Context,
	tx *gorm.DB,
	table string,
	tokenHash string,
	now time.Time,
) error {
	var row tokenRow
	if err := tx.WithContext(ctx).Table(table).
		Select("id", "token_expires_at", "token_used_at").
		Where("token_hash = ?", tokenHash).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("token: arama başarısız: %w", err)
	}

	if now.After(row.TokenExpiresAt) {
		return ErrTokenExpired
	}

	if row.TokenUsedAt != nil {
		return ErrTokenUsed
	}

	if err := tx.Table(table).
		Where("id = ?", row.ID).
		Update("token_used_at", now).Error; err != nil {
		return fmt.Errorf("token: kullanım işareti eklenemedi: %w", err)
	}

	return nil
}
