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

// ─── User Management Methods ───────────────────────────────────────────────────

// ListUsers returns a paginated, filtered list of users (admin only).
func (s *Service) ListUsers(ctx context.Context, filters UserFilters) (*UserListResponse, error) {
	users, total, err := s.repo.FindAllUsers(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("auth: list users: %w", err)
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	summaries := make([]UserSummary, 0, len(users))
	for _, u := range users {
		summaries = append(summaries, UserSummary{
			ID:        u.ID,
			FullName:  u.FullName,
			Email:     u.Email,
			Role:      string(u.Role),
			IsActive:  u.IsActive,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
		})
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &UserListResponse{
		Data:       summaries,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// CreateUser creates a new user account (admin only).
func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest, actorID string) (*User, error) {
	// Check if email already exists
	existing, _ := s.repo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already in use")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("auth: hash password: %w", err)
	}

	user := &User{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         UserRole(req.Role),
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("auth: create user: %w", err)
	}

	// Log the creation
	metadata, _ := json.Marshal(map[string]interface{}{
		"email": user.Email,
		"role":  user.Role,
	})
	_ = s.logRepo.Create(ctx, &logs.Log{
		EntityType: "user",
		EntityID:   user.ID,
		Action:     "user.created",
		ActorID:    &actorID,
		Metadata:   datatypes.JSON(metadata),
	})

	return user, nil
}

// GetUser returns a single user by ID (admin only).
func (s *Service) GetUser(ctx context.Context, id string) (*UserDetailResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("auth: get user: %w", err)
	}

	return &UserDetailResponse{
		ID:        user.ID,
		FullName:  user.FullName,
		Email:     user.Email,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateUser updates a user's information (admin only).
func (s *Service) UpdateUser(ctx context.Context, id string, req *UpdateUserRequest, actorID string) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrNotFound
		}
		return fmt.Errorf("auth: get user for update: %w", err)
	}

	// Build update map
	updates := make(map[string]interface{})
	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	if err := s.repo.UpdateFields(ctx, id, updates); err != nil {
		return fmt.Errorf("auth: update user: %w", err)
	}

	// Log the update
	metadata, _ := json.Marshal(map[string]interface{}{
		"user_id": id,
		"updates": updates,
	})
	_ = s.logRepo.Create(ctx, &logs.Log{
		EntityType: "user",
		EntityID:   id,
		Action:     "user.updated",
		ActorID:    &actorID,
		Metadata:   datatypes.JSON(metadata),
	})

	return nil
}

// ListActiveByRole returns active users filtered by role (for consultation member selection).
func (s *Service) ListActiveByRole(ctx context.Context, role string) ([]UserSummary, error) {
	filters := UserFilters{
		Role:     role,
		IsActive: boolPtr(true),
		PageSize: 100,
	}
	users, _, err := s.repo.FindAllUsers(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("auth: list active by role: %w", err)
	}

	summaries := make([]UserSummary, 0, len(users))
	for _, u := range users {
		summaries = append(summaries, UserSummary{
			ID:       u.ID,
			FullName: u.FullName,
			Email:    u.Email,
			Role:     string(u.Role),
			IsActive: u.IsActive,
		})
	}

	return summaries, nil
}

func boolPtr(b bool) *bool {
	return &b
}
