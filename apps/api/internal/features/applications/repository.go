package applications

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Repository handles all database operations for applications.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new applications repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new application record.
func (r *Repository) Create(ctx context.Context, app *Application) error {
	return r.db.WithContext(ctx).Create(app).Error
}

// FindByID loads a single application by primary key.
func (r *Repository) FindByID(ctx context.Context, id string) (*Application, error) {
	var app Application
	err := r.db.WithContext(ctx).First(&app, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// FindAll returns a paginated, filtered list of applications and the total count.
func (r *Repository) FindAll(ctx context.Context, filters ApplicationFilters) ([]*Application, int64, error) {
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	q := r.db.WithContext(ctx).Model(&Application{})

	if filters.MembershipType != "" {
		q = q.Where("membership_type = ?", filters.MembershipType)
	}
	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	}
	if filters.Search != "" {
		like := "%" + filters.Search + "%"
		q = q.Where("applicant_name LIKE ? OR applicant_email LIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("applications: count: %w", err)
	}

	var apps []*Application
	offset := (page - 1) * pageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&apps).Error; err != nil {
		return nil, 0, fmt.Errorf("applications: find all: %w", err)
	}

	return apps, total, nil
}

// UpdateStatus changes the status field of an application.
func (r *Repository) UpdateStatus(ctx context.Context, id string, status ApplicationStatus) error {
	result := r.db.WithContext(ctx).
		Table("applications").
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// FindByApplicantEmail returns all applications with the given email address.
func (r *Repository) FindByApplicantEmail(ctx context.Context, email string) ([]*Application, error) {
	var apps []*Application
	err := r.db.WithContext(ctx).
		Where("applicant_email = ?", email).
		Order("created_at DESC").
		Find(&apps).Error
	return apps, err
}

// GetTimeline returns all log entries for a specific application ordered by time.
func (r *Repository) GetTimeline(ctx context.Context, id string) ([]TimelineEntry, error) {
	type logEntry struct {
		Action    string `gorm:"column:action"`
		CreatedAt string `gorm:"column:created_at"`
		ActorRole string `gorm:"column:actor_role"`
		Notes     string `gorm:"column:notes"`
	}

	rows, err := r.db.WithContext(ctx).
		Raw(`SELECT action, created_at, actor_role, 
			IFNULL(JSON_UNQUOTE(JSON_EXTRACT(metadata, '$.notes')), '') AS notes
		FROM logs
		WHERE entity_type = 'application' AND entity_id = ?
		ORDER BY created_at ASC`, id).
		Rows()
	if err != nil {
		return nil, fmt.Errorf("applications: get timeline: %w", err)
	}
	defer rows.Close()

	var entries []TimelineEntry
	for rows.Next() {
		var e logEntry
		if err := r.db.ScanRows(rows, &e); err != nil {
			continue
		}
		entries = append(entries, TimelineEntry{
			Status:    ApplicationStatus(e.Action),
			ChangedBy: e.ActorRole,
			Notes:     e.Notes,
		})
	}
	return entries, nil
}

// CheckLinkedInUniqueness returns true if no other application has the given LinkedIn URL.
func (r *Repository) CheckLinkedInUniqueness(ctx context.Context, linkedInURL, excludeID string) (bool, error) {
	if linkedInURL == "" {
		return true, nil
	}
	var count int64
	q := r.db.WithContext(ctx).Model(&Application{}).
		Where("linked_in_url = ?", linkedInURL)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}
