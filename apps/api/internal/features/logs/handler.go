package logs

import (
	"encoding/json"
	"membership-system/api/internal/shared"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Handler handles HTTP requests for logs
type Handler struct {
	service *Service
	repo    *Repository
}

// NewHandler creates a new logs handler
func NewHandler(service *Service, repo *Repository) *Handler {
	return &Handler{
		service: service,
		repo:    repo,
	}
}

// ListFilters represents query parameters for filtering logs
type ListFilters struct {
	Action     string `query:"action"`
	EntityType string `query:"entity_type"`
	StartDate  string `query:"start_date"`
	EndDate    string `query:"end_date"`
	Page       int    `query:"page"`
	PageSize   int    `query:"page_size"`
}

// LogResponse represents the log entry returned to clients
type LogResponse struct {
	ID         string                 `json:"id"`
	ActorID    string                 `json:"actor_id"`
	ActorRole  string                 `json:"actor_role"`
	ActorName  *string                `json:"actor_name,omitempty"` // Only for admin
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// PaginatedLogsResponse represents the paginated response
type PaginatedLogsResponse struct {
	Data       []LogResponse `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

// List retrieves logs with filtering and pagination
// GET /api/v1/logs
func (h *Handler) List(c *fiber.Ctx) error {
	// Parse query parameters
	var filters ListFilters
	if err := c.QueryParser(&filters); err != nil {
		return shared.BadRequest(c, "Invalid query parameters")
	}

	// Default pagination
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 20
	}
	if filters.PageSize > 100 {
		filters.PageSize = 100
	}

	// Get user role from context (set by auth middleware)
	userRole, _ := c.Locals("userRole").(string)

	// Check if user has permission to view logs (yk, koordinator, or admin)
	if userRole != "yk" && userRole != "koordinator" && userRole != "admin" {
		return shared.Error(c, fiber.StatusForbidden, "FORBIDDEN", "You don't have permission to view logs")
	}

	// Only admin can see actor names
	isAdmin := userRole == "admin"

	// Build query
	query := h.repo.DB().WithContext(c.Context())

	// Apply filters
	if filters.Action != "" {
		query = query.Where("action LIKE ?", "%"+filters.Action+"%")
	}
	if filters.EntityType != "" {
		query = query.Where("entity_type = ?", filters.EntityType)
	}
	if filters.StartDate != "" {
		startTime, err := time.Parse("2006-01-02", filters.StartDate)
		if err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}
	if filters.EndDate != "" {
		endTime, err := time.Parse("2006-01-02", filters.EndDate)
		if err == nil {
			// Add one day to include the entire end date
			endTime = endTime.Add(24 * time.Hour)
			query = query.Where("created_at < ?", endTime)
		}
	}

	// Count total records
	var total int64
	if err := query.Model(&Log{}).Count(&total).Error; err != nil {
		return shared.Error(c, fiber.StatusInternalServerError, "DB_ERROR", "Failed to count logs")
	}

	// Calculate offset
	offset := (filters.Page - 1) * filters.PageSize

	// Fetch logs
	var logs []Log
	err := query.
		Order("created_at DESC").
		Limit(filters.PageSize).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		return shared.Error(c, fiber.StatusInternalServerError, "DB_ERROR", "Failed to retrieve logs")
	}

	// If admin, fetch actor names from users table
	var actorNames map[string]string
	if isAdmin && len(logs) > 0 {
		actorNames = make(map[string]string)
		var actorIDs []string
		for _, log := range logs {
			if log.ActorID != nil && *log.ActorID != "" {
				actorIDs = append(actorIDs, *log.ActorID)
			}
		}

		if len(actorIDs) > 0 {
			type UserName struct {
				ID   string
				Name string
			}
			var users []UserName
			h.repo.DB().Table("users").
				Select("id, CONCAT(first_name, ' ', last_name) as name").
				Where("id IN ?", actorIDs).
				Find(&users)

			for _, user := range users {
				actorNames[user.ID] = user.Name
			}
		}
	}

	// Convert to response format
	response := make([]LogResponse, len(logs))
	for i, log := range logs {
		actorID := ""
		if log.ActorID != nil {
			actorID = *log.ActorID
		}

		logResp := LogResponse{
			ID:         log.ID,
			ActorID:    actorID,
			ActorRole:  log.ActorRole,
			Action:     log.Action,
			EntityType: log.EntityType,
			EntityID:   log.EntityID,
			IPAddress:  log.IPAddress,
			CreatedAt:  log.CreatedAt,
		}

		// Add actor name only for admin
		if isAdmin && actorID != "" {
			if name, exists := actorNames[actorID]; exists {
				logResp.ActorName = &name
			}
		}

		// Parse metadata JSON
		if len(log.Metadata) > 0 {
			var metadata map[string]interface{}
			if err := json.Unmarshal(log.Metadata, &metadata); err == nil {
				logResp.Metadata = metadata
			}
		}

		response[i] = logResp
	}

	// Calculate total pages
	totalPages := int(total) / filters.PageSize
	if int(total)%filters.PageSize > 0 {
		totalPages++
	}

	return c.JSON(fiber.Map{
		"data": PaginatedLogsResponse{
			Data:       response,
			Total:      total,
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			TotalPages: totalPages,
		},
	})
}

// GetByID retrieves a single log by ID
// GET /api/v1/logs/:id
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return shared.BadRequest(c, "Log ID is required")
	}

	// Get user role from context
	userRole, _ := c.Locals("userRole").(string)

	// Check if user has permission to view logs
	if userRole != "yk" && userRole != "koordinator" && userRole != "admin" {
		return shared.Error(c, fiber.StatusForbidden, "FORBIDDEN", "You don't have permission to view logs")
	}

	isAdmin := userRole == "admin"

	// Fetch log
	var log Log
	err := h.repo.DB().WithContext(c.Context()).Where("id = ?", id).First(&log).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return shared.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Log not found")
		}
		return shared.Error(c, fiber.StatusInternalServerError, "DB_ERROR", "Failed to retrieve log")
	}

	// Convert to response
	actorID := ""
	if log.ActorID != nil {
		actorID = *log.ActorID
	}

	logResp := LogResponse{
		ID:         log.ID,
		ActorID:    actorID,
		ActorRole:  log.ActorRole,
		Action:     log.Action,
		EntityType: log.EntityType,
		EntityID:   log.EntityID,
		IPAddress:  log.IPAddress,
		CreatedAt:  log.CreatedAt,
	}

	// Add actor name only for admin
	if isAdmin && actorID != "" {
		type UserName struct {
			Name string
		}
		var userName UserName
		err := h.repo.DB().Table("users").
			Select("CONCAT(first_name, ' ', last_name) as name").
			Where("id = ?", actorID).
			First(&userName).Error
		if err == nil {
			logResp.ActorName = &userName.Name
		}
	}

	// Parse metadata JSON
	if len(log.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(log.Metadata, &metadata); err == nil {
			logResp.Metadata = metadata
		}
	}

	return c.JSON(fiber.Map{
		"data": logResp,
	})
}
