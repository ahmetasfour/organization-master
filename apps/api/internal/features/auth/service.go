package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"membership-system/api/internal/features/logs"
	"membership-system/api/internal/shared"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Service handles authentication business logic
type Service struct {
	repo             *Repository
	logRepo          *logs.Repository
	jwtSecret        string
	jwtRefreshSecret string
	accessTTL        time.Duration
	refreshTTL       time.Duration
}

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// NewService creates a new auth service
func NewService(repo *Repository, logRepo *logs.Repository, jwtSecret, jwtRefreshSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		repo:             repo,
		logRepo:          logRepo,
		jwtSecret:        jwtSecret,
		jwtRefreshSecret: jwtRefreshSecret,
		accessTTL:        accessTTL,
		refreshTTL:       refreshTTL,
	}
}

// Login authenticates a user and returns JWT tokens
func (s *Service) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	// Find user by email
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is disabled")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, shared.ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Log the login action
	if s.logRepo != nil {
		actorID := user.ID
		metadata, _ := json.Marshal(map[string]interface{}{"email": user.Email})
		_ = s.logRepo.Create(ctx, &logs.Log{
			EntityType: "user",
			EntityID:   user.ID,
			Action:     "auth.login",
			ActorID:    &actorID,
			ActorRole:  string(user.Role),
			Metadata:   datatypes.JSON(metadata),
		})
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserInfo{
			ID:       user.ID,
			FullName: user.FullName,
			Email:    user.Email,
			Role:     string(user.Role),
		},
	}, nil
}

// Refresh generates new tokens using a refresh token
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*RefreshResponse, error) {
	// Parse and validate refresh token
	token, err := jwt.ParseWithClaims(refreshToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtRefreshSecret), nil
	})

	if err != nil {
		return nil, shared.ErrUnauthorized
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, shared.ErrUnauthorized
	}

	// Find user
	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrUnauthorized
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if !user.IsActive {
		return nil, errors.New("account is disabled")
	}

	// Generate new tokens (token rotation)
	newAccessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		User: UserInfo{
			ID:       user.ID,
			FullName: user.FullName,
			Email:    user.Email,
			Role:     string(user.Role),
		},
	}, nil
}

// Logout logs out a user (stateless JWT - just log the event)
func (s *Service) Logout(ctx context.Context, userID string) error {
	// Find user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Log the logout action
	if s.logRepo != nil {
		actorID := userID
		_ = s.logRepo.Create(ctx, &logs.Log{
			EntityType: "user",
			EntityID:   userID,
			Action:     "auth.logout",
			ActorID:    &actorID,
			ActorRole:  string(user.Role),
		})
	}

	return nil
}

// generateAccessToken generates an access JWT token
func (s *Service) generateAccessToken(user *User) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// generateRefreshToken generates a refresh JWT token
func (s *Service) generateRefreshToken(user *User) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtRefreshSecret))
}

// ValidateToken validates and parses a JWT access token
func (s *Service) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, shared.ErrUnauthorized
	}

	return claims, nil
}
