package logs

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Log represents an immutable audit log entry
// This is an APPEND-ONLY table - no updates or deletions allowed
type Log struct {
	ID string `gorm:"type:char(36);primaryKey" json:"id"`

	// Action identifier (e.g., "application.created", "vote.cast", "auth.login")
	Action string `gorm:"type:varchar(255);not null;index" json:"action"`

	// Human-readable description (e.g., "YK Üyesi Ahmet Yılmaz başvuruya pozitif oy verdi")
	Description string `gorm:"type:varchar(500);index" json:"description,omitempty"`

	// Actor information
	ActorID    *string `gorm:"type:char(36);index" json:"actor_id,omitempty"`
	ActorRole  string  `gorm:"type:varchar(50)" json:"actor_role,omitempty"`
	ActorEmail string  `gorm:"type:varchar(255)" json:"actor_email,omitempty"`

	// Target entity
	EntityType string `gorm:"type:varchar(50);index:idx_entity_type_id" json:"entity_type,omitempty"`
	EntityID   string `gorm:"type:char(36);index:idx_entity_type_id" json:"entity_id,omitempty"`

	// Request metadata
	IPAddress string `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent string `gorm:"type:text" json:"user_agent,omitempty"`

	// Additional context as JSON
	Metadata datatypes.JSON `gorm:"type:json" json:"metadata,omitempty"`

	// Immutable timestamp
	CreatedAt time.Time `gorm:"not null;autoCreateTime;index" json:"created_at"`
}

// TableName specifies the table name for GORM
func (Log) TableName() string {
	return "logs"
}

// BeforeCreate GORM hook to generate UUID before creating a log
func (l *Log) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}

// IsSystemAction checks if this is a system-level action
func (l *Log) IsSystemAction() bool {
	return l.ActorID == nil
}

// HasMetadata checks if the log has additional metadata
func (l *Log) HasMetadata() bool {
	return len(l.Metadata) > 0
}
