package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/zen/shared/pkg/tenant_models"
	"project-service/internal/repositories"
)

type ProjectService interface {
	// Project CRUD
	CreateProject(userID, tenantID string, req *tenant_models.ProjectCreateRequest) (*tenant_models.ProjectResponse, error)
	GetProject(userID, tenantID, projectID string) (*tenant_models.ProjectResponse, error)
	UpdateProject(userID, tenantID, projectID string, req *tenant_models.ProjectUpdateRequest) (*tenant_models.ProjectResponse, error)
	DeleteProject(userID, tenantID, projectID string) error
	ListProjects(userID, tenantID string, limit, offset int, filters repositories.ProjectFilters) ([]*tenant_models.ProjectResponse, int64, error)

	// Project members
	AddProjectMember(userID, tenantID, projectID string, req *tenant_models.ProjectMemberCreateRequest) (*tenant_models.ProjectMemberResponse, error)
	RemoveProjectMember(userID, tenantID, projectID, memberUserID string) error
	ListProjectMembers(userID, tenantID, projectID string) ([]*tenant_models.ProjectMemberResponse, error)
	UpdateProjectMemberRole(userID, tenantID, projectID, memberUserID string, req *tenant_models.ProjectMemberUpdateRequest) (*tenant_models.ProjectMemberResponse, error)

	// Project stats
	GetProjectStats(userID, tenantID, projectID string, dateFrom, dateTo *time.Time) (*repositories.ProjectStats, error)
	GetUserProjectStats(userID, tenantID string, dateFrom, dateTo *time.Time) (*repositories.UserProjectStats, error)
}

type projectService struct {
	repo   repositories.ProjectRepository
	logger *zap.Logger
}

func NewProjectService(repo repositories.ProjectRepository, logger *zap.Logger) ProjectService {
	return &projectService{
		repo:   repo,
		logger: logger,
	}
}

func (s *projectService) CreateProject(userID, tenantID string, req *tenant_models.ProjectCreateRequest) (*tenant_models.ProjectResponse, error) {
	// Validate input
	if req.Name == "" {
		return nil, errors.New("project name is required")
	}
	if req.Key == "" {
		return nil, errors.New("project key is required")
	}

	// Create project model
	project := &tenant_models.Project{
		ID:          uuid.New().String(),
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		ProjectType: req.ProjectType,
		Methodology: req.Methodology,
		Status:      tenant_models.ProjectStatusActive, // Default to active
		LeadID:      req.LeadID,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set defaults if not provided
	if project.ProjectType == "" {
		project.ProjectType = tenant_models.ProjectTypeSoftware
	}
	if project.Methodology == "" {
		project.Methodology = tenant_models.MethodologyKanban
	}

	// Save project
	err := s.repo.CreateProject(tenantID, project)
	if err != nil {
		s.logger.Error("Failed to create project", zap.Error(err))
		return nil, err
	}

	// Add lead as project member
	member := &tenant_models.ProjectMember{
		ID:        uuid.New().String(),
		ProjectID: project.ID,
		UserID:    req.LeadID,
		Role:      "lead",
		AddedAt:   time.Now(),
	}
	
	if err := s.repo.AddProjectMember(tenantID, member); err != nil {
		s.logger.Warn("Failed to add lead as project member", zap.Error(err))
	}

	response := project.ToResponse()
	return &response, nil
}

func (s *projectService) GetProject(userID, tenantID, projectID string) (*tenant_models.ProjectResponse, error) {
	project, err := s.repo.GetProject(tenantID, projectID)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this project
	if !s.userCanAccessProject(userID, tenantID, projectID) {
		return nil, errors.New("access denied")
	}

	response := project.ToResponse()
	return &response, nil
}

func (s *projectService) UpdateProject(userID, tenantID, projectID string, req *tenant_models.ProjectUpdateRequest) (*tenant_models.ProjectResponse, error) {
	// Get existing project
	project, err := s.repo.GetProject(tenantID, projectID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if !s.userCanModifyProject(userID, tenantID, projectID) {
		return nil, errors.New("access denied")
	}

	// Update fields
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.Status != nil {
		project.Status = *req.Status
	}
	if req.ProjectType != nil {
		project.ProjectType = *req.ProjectType
	}
	if req.Methodology != nil {
		project.Methodology = *req.Methodology
	}
	if req.LeadID != nil {
		project.LeadID = *req.LeadID
	}

	project.UpdatedAt = time.Now()

	// Save
	err = s.repo.UpdateProject(tenantID, project)
	if err != nil {
		return nil, err
	}

	response := project.ToResponse()
	return &response, nil
}

func (s *projectService) DeleteProject(userID, tenantID, projectID string) error {
	// Check permissions
	if !s.userCanModifyProject(userID, tenantID, projectID) {
		return errors.New("access denied")
	}

	return s.repo.DeleteProject(tenantID, projectID)
}

func (s *projectService) ListProjects(userID, tenantID string, limit, offset int, filters repositories.ProjectFilters) ([]*tenant_models.ProjectResponse, int64, error) {
	projects, total, err := s.repo.ListProjects(tenantID, limit, offset, filters)
	if err != nil {
		return nil, 0, err
	}

	var responses []*tenant_models.ProjectResponse
	for _, project := range projects {
		response := project.ToResponse()
		responses = append(responses, &response)
	}

	return responses, total, nil
}

func (s *projectService) AddProjectMember(userID, tenantID, projectID string, req *tenant_models.ProjectMemberCreateRequest) (*tenant_models.ProjectMemberResponse, error) {
	// Check permissions
	if !s.userCanModifyProject(userID, tenantID, projectID) {
		return nil, errors.New("access denied")
	}

	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = tenant_models.ProjectRoleMember
	}

	member := &tenant_models.ProjectMember{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      role,
		AddedAt:   time.Now(),
		AddedBy:   userID,
	}

	err := s.repo.AddProjectMember(tenantID, member)
	if err != nil {
		return nil, err
	}

	response := member.ToResponse()
	return &response, nil
}

func (s *projectService) RemoveProjectMember(userID, tenantID, projectID, memberUserID string) error {
	// Check permissions
	if !s.userCanModifyProject(userID, tenantID, projectID) {
		return errors.New("access denied")
	}

	// Don't allow removing the project lead
	project, err := s.repo.GetProject(tenantID, projectID)
	if err != nil {
		return err
	}
	if project.LeadID == memberUserID {
		return errors.New("cannot remove project lead")
	}

	return s.repo.RemoveProjectMember(tenantID, projectID, memberUserID)
}

func (s *projectService) ListProjectMembers(userID, tenantID, projectID string) ([]*tenant_models.ProjectMemberResponse, error) {
	// Check access
	if !s.userCanAccessProject(userID, tenantID, projectID) {
		return nil, errors.New("access denied")
	}

	members, err := s.repo.ListProjectMembers(tenantID, projectID)
	if err != nil {
		return nil, err
	}

	var responses []*tenant_models.ProjectMemberResponse
	for _, member := range members {
		response := member.ToResponse()
		responses = append(responses, &response)
	}

	return responses, nil
}

func (s *projectService) UpdateProjectMemberRole(userID, tenantID, projectID, memberUserID string, req *tenant_models.ProjectMemberUpdateRequest) (*tenant_models.ProjectMemberResponse, error) {
	// Check permissions
	if !s.userCanModifyProject(userID, tenantID, projectID) {
		return nil, errors.New("access denied")
	}

	if req.Role == nil {
		return nil, errors.New("role is required")
	}

	err := s.repo.UpdateProjectMemberRole(tenantID, projectID, memberUserID, string(*req.Role))
	if err != nil {
		return nil, err
	}

	member, err := s.repo.GetProjectMember(tenantID, projectID, memberUserID)
	if err != nil {
		return nil, err
	}

	response := member.ToResponse()
	return &response, nil
}

func (s *projectService) GetProjectStats(userID, tenantID, projectID string, dateFrom, dateTo *time.Time) (*repositories.ProjectStats, error) {
	// Check access
	if !s.userCanAccessProject(userID, tenantID, projectID) {
		return nil, errors.New("access denied")
	}

	return s.repo.GetProjectStats(tenantID, projectID, dateFrom, dateTo)
}

func (s *projectService) GetUserProjectStats(userID, tenantID string, dateFrom, dateTo *time.Time) (*repositories.UserProjectStats, error) {
	return s.repo.GetUserProjectStats(tenantID, userID, dateFrom, dateTo)
}


func (s *projectService) userCanAccessProject(userID, tenantID, projectID string) bool {
	// Check if user is project lead or creator
	project, err := s.repo.GetProject(tenantID, projectID)
	if err != nil {
		return false
	}
	if project.LeadID == userID || project.CreatedBy == userID {
		return true
	}

	// Check if user is project member
	_, err = s.repo.GetProjectMember(tenantID, projectID, userID)
	return err == nil
}

func (s *projectService) userCanModifyProject(userID, tenantID, projectID string) bool {
	// Check if user is project lead or creator
	project, err := s.repo.GetProject(tenantID, projectID)
	if err != nil {
		return false
	}
	if project.LeadID == userID || project.CreatedBy == userID {
		return true
	}

	// Check if user is admin member
	member, err := s.repo.GetProjectMember(tenantID, projectID, userID)
	if err != nil {
		return false
	}
	return member.Role == "admin" || member.Role == "lead"
}