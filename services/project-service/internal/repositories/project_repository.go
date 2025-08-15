package repositories

import (
	"errors"
	"time"

	"github.com/zen/shared/pkg/database"
	"github.com/zen/shared/pkg/tenant_models"
)

type ProjectRepository interface {
	// Project CRUD
	CreateProject(tenantID string, project *tenant_models.Project) error
	GetProject(tenantID, projectID string) (*tenant_models.Project, error)
	UpdateProject(tenantID string, project *tenant_models.Project) error
	DeleteProject(tenantID, projectID string) error
	ListProjects(tenantID string, limit, offset int, filters ProjectFilters) ([]*tenant_models.Project, int64, error)

	// Project members
	AddProjectMember(tenantID string, member *tenant_models.ProjectMember) error
	RemoveProjectMember(tenantID, projectID, userID string) error
	ListProjectMembers(tenantID, projectID string) ([]*tenant_models.ProjectMember, error)
	GetProjectMember(tenantID, projectID, userID string) (*tenant_models.ProjectMember, error)
	UpdateProjectMemberRole(tenantID, projectID, userID, role string) error

	// Project stats
	GetProjectStats(tenantID, projectID string, dateFrom, dateTo *time.Time) (*ProjectStats, error)
	GetUserProjectStats(tenantID, userID string, dateFrom, dateTo *time.Time) (*UserProjectStats, error)
}

type ProjectFilters struct {
	Status      string
	LeadID      string
	CreatedBy   string
	MemberID    string
	Search      string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type ProjectStats struct {
	TotalProjects    int64              `json:"total_projects"`
	ActiveProjects   int64              `json:"active_projects"`
	CompletedProjects int64             `json:"completed_projects"`
	OnHoldProjects   int64              `json:"on_hold_projects"`
	StatusBreakdown  map[string]int64   `json:"status_breakdown"`
	MonthlyBreakdown []MonthlyProjectStats `json:"monthly_breakdown"`
}

type MonthlyProjectStats struct {
	Month     string `json:"month"`
	Year      int    `json:"year"`
	Created   int64  `json:"created"`
	Completed int64  `json:"completed"`
}

type UserProjectStats struct {
	TotalProjects     int64 `json:"total_projects"`
	ProjectsAsOwner   int64 `json:"projects_as_owner"`
	ProjectsAsMember  int64 `json:"projects_as_member"`
	ActiveProjects    int64 `json:"active_projects"`
	CompletedProjects int64 `json:"completed_projects"`
}

type projectRepository struct {
	tenantDBManager *database.TenantDatabaseManager
}

func NewProjectRepository(tenantDBManager *database.TenantDatabaseManager) ProjectRepository {
	return &projectRepository{
		tenantDBManager: tenantDBManager,
	}
}

func (r *projectRepository) CreateProject(tenantID string, project *tenant_models.Project) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	return db.Create(project).Error
}

func (r *projectRepository) GetProject(tenantID, projectID string) (*tenant_models.Project, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var project tenant_models.Project
	err = db.Preload("Members").First(&project, "id = ?", projectID).Error
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (r *projectRepository) UpdateProject(tenantID string, project *tenant_models.Project) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	project.UpdatedAt = time.Now()
	return db.Save(project).Error
}

func (r *projectRepository) DeleteProject(tenantID, projectID string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	// Soft delete the project
	return db.Delete(&tenant_models.Project{}, "id = ?", projectID).Error
}

func (r *projectRepository) ListProjects(tenantID string, limit, offset int, filters ProjectFilters) ([]*tenant_models.Project, int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, 0, err
	}

	var projects []*tenant_models.Project
	var total int64

	query := db.Model(&tenant_models.Project{})

	// Apply filters
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.LeadID != "" {
		query = query.Where("lead_id = ?", filters.LeadID)
	}
	if filters.CreatedBy != "" {
		query = query.Where("created_by = ?", filters.CreatedBy)
	}
	if filters.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}
	if filters.CreatedFrom != nil {
		query = query.Where("created_at >= ?", filters.CreatedFrom)
	}
	if filters.CreatedTo != nil {
		query = query.Where("created_at <= ?", filters.CreatedTo)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get projects with members preloaded
	err = query.Preload("Members").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&projects).Error

	return projects, total, err
}

func (r *projectRepository) AddProjectMember(tenantID string, member *tenant_models.ProjectMember) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	// Check if member already exists
	var existing tenant_models.ProjectMember
	err = db.Where("project_id = ? AND user_id = ?", member.ProjectID, member.UserID).First(&existing).Error
	if err == nil {
		return errors.New("user is already a member of this project")
	}

	return db.Create(member).Error
}

func (r *projectRepository) RemoveProjectMember(tenantID, projectID, userID string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	return db.Delete(&tenant_models.ProjectMember{}, "project_id = ? AND user_id = ?", projectID, userID).Error
}

func (r *projectRepository) ListProjectMembers(tenantID, projectID string) ([]*tenant_models.ProjectMember, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var members []*tenant_models.ProjectMember
	err = db.Where("project_id = ?", projectID).Find(&members).Error
	return members, err
}

func (r *projectRepository) GetProjectMember(tenantID, projectID, userID string) (*tenant_models.ProjectMember, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var member tenant_models.ProjectMember
	err = db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error
	return &member, err
}

func (r *projectRepository) UpdateProjectMemberRole(tenantID, projectID, userID, role string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	return db.Model(&tenant_models.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Update("role", role).Error
}

func (r *projectRepository) GetProjectStats(tenantID, projectID string, dateFrom, dateTo *time.Time) (*ProjectStats, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	stats := &ProjectStats{
		StatusBreakdown: make(map[string]int64),
	}

	query := db.Model(&tenant_models.Project{})
	if projectID != "" {
		query = query.Where("id = ?", projectID)
	}
	if dateFrom != nil {
		query = query.Where("created_at >= ?", dateFrom)
	}
	if dateTo != nil {
		query = query.Where("created_at <= ?", dateTo)
	}

	// Total projects
	query.Count(&stats.TotalProjects)

	// Status breakdown
	var statusCounts []struct {
		Status string
		Count  int64
	}
	db.Model(&tenant_models.Project{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusCounts)

	for _, sc := range statusCounts {
		stats.StatusBreakdown[sc.Status] = sc.Count
		switch sc.Status {
		case "active":
			stats.ActiveProjects = sc.Count
		case "completed":
			stats.CompletedProjects = sc.Count
		case "on_hold":
			stats.OnHoldProjects = sc.Count
		}
	}

	return stats, nil
}

func (r *projectRepository) GetUserProjectStats(tenantID, userID string, dateFrom, dateTo *time.Time) (*UserProjectStats, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	stats := &UserProjectStats{}

	// Projects as lead
	query := db.Model(&tenant_models.Project{}).Where("lead_id = ?", userID)
	if dateFrom != nil {
		query = query.Where("created_at >= ?", dateFrom)
	}
	if dateTo != nil {
		query = query.Where("created_at <= ?", dateTo)
	}
	query.Count(&stats.ProjectsAsOwner)  // Using ProjectsAsOwner field for lead projects

	// Projects as member (including owner)
	memberQuery := db.Model(&tenant_models.ProjectMember{}).Where("user_id = ?", userID)
	if dateFrom != nil {
		memberQuery = memberQuery.Where("created_at >= ?", dateFrom)
	}
	if dateTo != nil {
		memberQuery = memberQuery.Where("created_at <= ?", dateTo)
	}
	memberQuery.Count(&stats.ProjectsAsMember)

	stats.TotalProjects = stats.ProjectsAsOwner + stats.ProjectsAsMember

	return stats, nil
}