package auth

// LoginRequest represents the login credentials
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	User         UserInfo `json:"user"`
}

// UserInfo represents user information returned in login/refresh responses
type UserInfo struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// RefreshRequest represents the refresh token request
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// RefreshResponse is the same as LoginResponse
type RefreshResponse = LoginResponse

// ─── User Management DTOs ──────────────────────────────────────────────────────

// CreateUserRequest is the payload for POST /api/v1/users (admin only).
type CreateUserRequest struct {
	FullName string `json:"full_name" validate:"required,min=2,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"required,oneof=admin yk yik koordinator asil_uye"`
}

// UpdateUserRequest is the payload for PATCH /api/v1/users/:id (admin only).
type UpdateUserRequest struct {
	FullName string `json:"full_name" validate:"omitempty,min=2,max=255"`
	Role     string `json:"role" validate:"omitempty,oneof=admin yk yik koordinator asil_uye"`
	IsActive *bool  `json:"is_active"`
}

// UserFilters holds query parameters for filtering user lists.
type UserFilters struct {
	Role     string `query:"role"`
	IsActive *bool  `query:"is_active"`
	Search   string `query:"search"`
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
}

// UserSummary is used in list responses.
type UserSummary struct {
	ID        string `json:"id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// UserListResponse wraps a paginated user list.
type UserListResponse struct {
	Data       []UserSummary `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

// UserDetailResponse is the full user payload returned on GET /:id.
type UserDetailResponse struct {
	ID        string `json:"id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
