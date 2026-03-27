package logs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Service handles business logic for audit logging
type Service struct {
	repo *Repository
}

// NewService creates a new logs service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new audit log entry
func (s *Service) Create(ctx context.Context, actorID, actorRole, action, entityType, entityID string, metadata map[string]interface{}, ipAddress string) error {
	// Convert metadata to JSON
	var metadataJSON []byte
	if metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(metadata)
		if err != nil {
			// Log error but don't fail the operation
			metadataJSON = []byte("{}")
		}
	} else {
		metadataJSON = []byte("{}")
	}

	log := &Log{
		ID:         uuid.New().String(),
		Action:     action,
		ActorID:    &actorID,
		ActorRole:  actorRole,
		EntityType: entityType,
		EntityID:   entityID,
		IPAddress:  ipAddress,
		Metadata:   metadataJSON,
		CreatedAt:  time.Now(),
	}

	return s.repo.Create(ctx, log)
}

// FindByEntityID retrieves logs for a specific entity
func (s *Service) FindByEntityID(ctx context.Context, entityType, entityID string) ([]Log, error) {
	return s.repo.FindByEntityID(ctx, entityType, entityID)
}

// FindByAction retrieves logs for a specific action
func (s *Service) FindByAction(ctx context.Context, action string) ([]Log, error) {
	return s.repo.FindByAction(ctx, action)
}
