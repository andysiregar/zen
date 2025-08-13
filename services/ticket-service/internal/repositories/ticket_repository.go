package repositories

import (
	"fmt"
	"time"

	"github.com/zen/shared/pkg/database"
	"github.com/zen/shared/pkg/models"
	"gorm.io/gorm"
)

type TicketRepository interface {
	// Ticket CRUD
	Create(tenantID string, ticket *models.Ticket) error
	GetByID(tenantID, ticketID string) (*models.Ticket, error)
	Update(tenantID string, ticket *models.Ticket) error
	Delete(tenantID, ticketID string) error
	List(tenantID string, limit, offset int, filters TicketFilters) ([]*models.Ticket, int64, error)
	
	// Ticket search and filtering
	Search(tenantID, query string, limit, offset int) ([]*models.Ticket, int64, error)
	GetByStatus(tenantID string, status models.TicketStatus, limit, offset int) ([]*models.Ticket, error)
	GetByAssignee(tenantID, assigneeID string, limit, offset int) ([]*models.Ticket, error)
	GetByReporter(tenantID, reporterID string, limit, offset int) ([]*models.Ticket, error)
	GetByProject(tenantID, projectID string, limit, offset int) ([]*models.Ticket, error)
	
	// Comments
	CreateComment(tenantID string, comment *models.TicketComment) error
	GetComments(tenantID, ticketID string, includeInternal bool) ([]*models.TicketComment, error)
	UpdateComment(tenantID string, comment *models.TicketComment) error
	DeleteComment(tenantID, commentID string) error
	
	// Attachments
	CreateAttachment(tenantID string, attachment *models.TicketAttachment) error
	GetAttachments(tenantID, ticketID string) ([]*models.TicketAttachment, error)
	DeleteAttachment(tenantID, attachmentID string) error
	
	// Statistics
	GetTicketStats(tenantID string, dateFrom, dateTo *time.Time) (TicketStats, error)
}

type TicketFilters struct {
	Status     string
	Priority   string
	Type       string
	AssigneeID string
	ReporterID string
	ProjectID  string
	Category   string
	Tags       []string
	DateFrom   *time.Time
	DateTo     *time.Time
}

type TicketStats struct {
	TotalTickets       int64            `json:"total_tickets"`
	OpenTickets        int64            `json:"open_tickets"`
	InProgressTickets  int64            `json:"in_progress_tickets"`
	ResolvedTickets    int64            `json:"resolved_tickets"`
	ClosedTickets      int64            `json:"closed_tickets"`
	ByStatus           map[string]int64 `json:"by_status"`
	ByPriority         map[string]int64 `json:"by_priority"`
	ByType             map[string]int64 `json:"by_type"`
	AverageResolutionTime time.Duration `json:"average_resolution_time"`
}

type ticketRepository struct {
	dbManager *database.DatabaseManager
}

func NewTicketRepository(dbManager *database.DatabaseManager) TicketRepository {
	return &ticketRepository{
		dbManager: dbManager,
	}
}

func (r *ticketRepository) getTenantDB(tenantID string) (*gorm.DB, error) {
	// Get tenant connection info from master database
	var tenant models.Tenant
	err := r.dbManager.GetMasterDB().Where("id = ?", tenantID).First(&tenant).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant info: %w", err)
	}

	// Create TenantConnectionInfo
	connInfo := database.TenantConnectionInfo{
		TenantID:          tenantID,
		Host:              tenant.DbHost,
		Port:              tenant.DbPort,
		User:              tenant.DbUser,
		EncryptedPassword: tenant.DbPasswordEncrypted,
		DBName:            tenant.DbName,
		SSLMode:           tenant.DbSslMode,
	}

	return r.dbManager.GetTenantDB(connInfo)
}

func (r *ticketRepository) Create(tenantID string, ticket *models.Ticket) error {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	ticket.TenantID = tenantID
	return db.Create(ticket).Error
}

func (r *ticketRepository) GetByID(tenantID, ticketID string) (*models.Ticket, error) {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	var ticket models.Ticket
	err = db.Where("id = ? AND tenant_id = ?", ticketID, tenantID).First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) Update(tenantID string, ticket *models.Ticket) error {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	return db.Where("tenant_id = ?", tenantID).Save(ticket).Error
}

func (r *ticketRepository) Delete(tenantID, ticketID string) error {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	return db.Where("id = ? AND tenant_id = ?", ticketID, tenantID).Delete(&models.Ticket{}).Error
}

func (r *ticketRepository) List(tenantID string, limit, offset int, filters TicketFilters) ([]*models.Ticket, int64, error) {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	query := db.Where("tenant_id = ?", tenantID)
	
	// Apply filters
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Priority != "" {
		query = query.Where("priority = ?", filters.Priority)
	}
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	if filters.AssigneeID != "" {
		query = query.Where("assignee_id = ?", filters.AssigneeID)
	}
	if filters.ReporterID != "" {
		query = query.Where("reporter_id = ?", filters.ReporterID)
	}
	if filters.ProjectID != "" {
		query = query.Where("project_id = ?", filters.ProjectID)
	}
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}
	if filters.DateFrom != nil {
		query = query.Where("created_at >= ?", filters.DateFrom)
	}
	if filters.DateTo != nil {
		query = query.Where("created_at <= ?", filters.DateTo)
	}
	
	// Get total count
	var total int64
	countQuery := query
	err = countQuery.Model(&models.Ticket{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get tickets
	var tickets []*models.Ticket
	err = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tickets).Error
	
	return tickets, total, err
}

func (r *ticketRepository) Search(tenantID, query string, limit, offset int) ([]*models.Ticket, int64, error) {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	searchQuery := db.Where("tenant_id = ?", tenantID).
		Where("title ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")
	
	var total int64
	err = searchQuery.Model(&models.Ticket{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	var tickets []*models.Ticket
	err = searchQuery.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tickets).Error
	
	return tickets, total, err
}

func (r *ticketRepository) GetByStatus(tenantID string, status models.TicketStatus, limit, offset int) ([]*models.Ticket, error) {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	var tickets []*models.Ticket
	err = db.Where("tenant_id = ? AND status = ?", tenantID, status).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&tickets).Error
	
	return tickets, err
}

func (r *ticketRepository) GetByAssignee(tenantID, assigneeID string, limit, offset int) ([]*models.Ticket, error) {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	var tickets []*models.Ticket
	err = db.Where("tenant_id = ? AND assignee_id = ?", tenantID, assigneeID).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&tickets).Error
	
	return tickets, err
}

func (r *ticketRepository) GetByReporter(tenantID, reporterID string, limit, offset int) ([]*models.Ticket, error) {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	var tickets []*models.Ticket
	err = db.Where("tenant_id = ? AND reporter_id = ?", tenantID, reporterID).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&tickets).Error
	
	return tickets, err
}

func (r *ticketRepository) GetByProject(tenantID, projectID string, limit, offset int) ([]*models.Ticket, error) {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	var tickets []*models.Ticket
	err = db.Where("tenant_id = ? AND project_id = ?", tenantID, projectID).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&tickets).Error
	
	return tickets, err
}

// Comment methods
func (r *ticketRepository) CreateComment(tenantID string, comment *models.TicketComment) error {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	return db.Create(comment).Error
}

func (r *ticketRepository) GetComments(tenantID, ticketID string, includeInternal bool) ([]*models.TicketComment, error) {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	query := db.Where("ticket_id = ?", ticketID)
	if !includeInternal {
		query = query.Where("is_internal = ?", false)
	}
	
	var comments []*models.TicketComment
	err = query.Order("created_at ASC").Find(&comments).Error
	
	return comments, err
}

func (r *ticketRepository) UpdateComment(tenantID string, comment *models.TicketComment) error {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	return db.Save(comment).Error
}

func (r *ticketRepository) DeleteComment(tenantID, commentID string) error {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	return db.Where("id = ?", commentID).Delete(&models.TicketComment{}).Error
}

// Attachment methods
func (r *ticketRepository) CreateAttachment(tenantID string, attachment *models.TicketAttachment) error {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	return db.Create(attachment).Error
}

func (r *ticketRepository) GetAttachments(tenantID, ticketID string) ([]*models.TicketAttachment, error) {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	var attachments []*models.TicketAttachment
	err = db.Where("ticket_id = ?", ticketID).Order("created_at ASC").Find(&attachments).Error
	
	return attachments, err
}

func (r *ticketRepository) DeleteAttachment(tenantID, attachmentID string) error {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	return db.Where("id = ?", attachmentID).Delete(&models.TicketAttachment{}).Error
}

// Statistics
func (r *ticketRepository) GetTicketStats(tenantID string, dateFrom, dateTo *time.Time) (TicketStats, error) {
	db, err := r.getTenantDB(tenantID)
	if err != nil {
		return TicketStats{}, fmt.Errorf("failed to get tenant database: %w", err)
	}
	
	var stats TicketStats
	
	query := db.Where("tenant_id = ?", tenantID)
	if dateFrom != nil {
		query = query.Where("created_at >= ?", dateFrom)
	}
	if dateTo != nil {
		query = query.Where("created_at <= ?", dateTo)
	}
	
	// Total tickets
	query.Model(&models.Ticket{}).Count(&stats.TotalTickets)
	
	// By status
	stats.ByStatus = make(map[string]int64)
	var statusCounts []struct {
		Status string
		Count  int64
	}
	query.Model(&models.Ticket{}).Select("status, COUNT(*) as count").Group("status").Find(&statusCounts)
	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
		
		// Set individual status counts
		switch sc.Status {
		case string(models.TicketStatusOpen):
			stats.OpenTickets = sc.Count
		case string(models.TicketStatusInProgress):
			stats.InProgressTickets = sc.Count
		case string(models.TicketStatusResolved):
			stats.ResolvedTickets = sc.Count
		case string(models.TicketStatusClosed):
			stats.ClosedTickets = sc.Count
		}
	}
	
	// By priority
	stats.ByPriority = make(map[string]int64)
	var priorityCounts []struct {
		Priority string
		Count    int64
	}
	query.Model(&models.Ticket{}).Select("priority, COUNT(*) as count").Group("priority").Find(&priorityCounts)
	for _, pc := range priorityCounts {
		stats.ByPriority[pc.Priority] = pc.Count
	}
	
	// By type
	stats.ByType = make(map[string]int64)
	var typeCounts []struct {
		Type  string
		Count int64
	}
	query.Model(&models.Ticket{}).Select("type, COUNT(*) as count").Group("type").Find(&typeCounts)
	for _, tc := range typeCounts {
		stats.ByType[tc.Type] = tc.Count
	}
	
	// TODO: Calculate average resolution time
	stats.AverageResolutionTime = 0
	
	return stats, nil
}