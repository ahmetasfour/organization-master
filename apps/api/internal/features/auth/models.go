package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents all possible user roles in the system
type UserRole string

const (
	RoleAdmin       UserRole = "admin"
	RoleKoordinator UserRole = "koordinator"
	RoleYK          UserRole = "yk"
	RoleYIK         UserRole = "yik"
	RoleAsilUye     UserRole = "asil_uye"
	RoleYIKUye      UserRole = "yik_uye"
)

// User represents a system user with authentication and role information
type User struct {
	ID           string    `gorm:"type:char(36);primaryKey" json:"id"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	FullName     string    `gorm:"type:varchar(255);not null" json:"full_name"`
	Role         UserRole  `gorm:"type:enum('admin','koordinator','yk','yik','asil_uye','yik_uye');not null" json:"role"`
	IsActive     bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt    time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM hook to generate UUID before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// IsSystemUser checks if the user has system-level privileges
func (u *User) IsSystemUser() bool {
	return u.Role == RoleAdmin || u.Role == RoleKoordinator
}

// CanVote checks if the user has voting privileges
func (u *User) CanVote() bool {
	return u.Role == RoleYK || u.Role == RoleYIK
}

// CanProposeApplication checks if the user can propose new applications
func (u *User) CanProposeApplication() bool {
	return u.Role == RoleAsilUye || u.Role == RoleYIKUye
}
