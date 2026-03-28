package auth

import (
	"context"

	"gorm.io/gorm"
)

// Repository handles database operations for auth
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new auth repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// FindByEmail finds a user by email address
func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID finds a user by ID
func (r *Repository) FindByID(ctx context.Context, id string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *Repository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update updates a user
func (r *Repository) Update(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdateFields updates specific fields of a user
func (r *Repository) UpdateFields(ctx context.Context, id string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(updates).Error
}

// FindAllUsers returns a paginated, filtered list of users
func (r *Repository) FindAllUsers(ctx context.Context, filters UserFilters) ([]*User, int64, error) {
	query := r.db.WithContext(ctx).Model(&User{})

	// Apply filters
	if filters.Role != "" {
		query = query.Where("role = ?", filters.Role)
	}
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}
	if filters.Search != "" {
		search := "%" + filters.Search + "%"
		query = query.Where("full_name LIKE ? OR email LIKE ?", search, search)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var users []*User
	err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&users).Error
	return users, total, err
}
