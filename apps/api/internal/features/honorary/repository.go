package honorary

import (
	"context"

	"membership-system/api/internal/features/applications"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, application *applications.Application) error {
	return r.db.WithContext(ctx).Create(application).Error
}

func (r *Repository) FindAll(ctx context.Context) ([]*ProposalResponse, error) {
	var results []*ProposalResponse

	err := r.db.WithContext(ctx).
		Table("applications a").
		Select(`
			a.id as application_id,
			a.applicant_name as nominee_name,
			a.linked_in_url as nominee_linkedin,
			a.proposal_reason,
			u.full_name as proposed_by,
			a.status,
			a.created_at
		`).
		Joins("LEFT JOIN users u ON a.proposed_by_user_id = u.id").
		Where("a.membership_type = ?", "onursal").
		Order("a.created_at DESC").
		Scan(&results).Error

	return results, err
}

func (r *Repository) CheckLinkedInExists(ctx context.Context, linkedinURL string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&applications.Application{}).
		Where("linked_in_url = ?", linkedinURL).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) GetYKMembers(ctx context.Context) ([]struct {
	ID       string
	FullName string
	Email    string
}, error) {
	var members []struct {
		ID       string
		FullName string
		Email    string
	}

	err := r.db.WithContext(ctx).
		Table("users").
		Select("id, full_name, email").
		Where("role = ?", "yk").
		Scan(&members).Error

	return members, err
}
