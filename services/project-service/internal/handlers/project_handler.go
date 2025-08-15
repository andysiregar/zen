package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zen/shared/pkg/middleware"
	"github.com/zen/shared/pkg/tenant_models"
	"github.com/zen/shared/pkg/utils"
	"project-service/internal/repositories"
	"project-service/internal/services"
)

type ProjectHandler struct {
	service services.ProjectService
	logger  *zap.Logger
}

func NewProjectHandler(service services.ProjectService, logger *zap.Logger) *ProjectHandler {
	return &ProjectHandler{
		service: service,
		logger:  logger,
	}
}

// Helper function to get user and tenant context
func (h *ProjectHandler) getUserAndTenantContext(c *gin.Context) (string, string, error) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return "", "", fmt.Errorf("user ID not found in context")
	}

	tenantContext, err := middleware.GetTenantContext(c)
	if err != nil {
		return "", "", fmt.Errorf("tenant context not found: %w", err)
	}

	return userID, tenantContext.TenantID, nil
}

// CreateProject handles POST /projects
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	var req tenant_models.ProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	project, err := h.service.CreateProject(userID, tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to create project", zap.Error(err))
		utils.InternalServerErrorResponse(c, "Failed to create project")
		return
	}

	utils.CreatedResponse(c, project, "Project created successfully")
}

// GetProject handles GET /projects/:id
func (h *ProjectHandler) GetProject(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		utils.BadRequestResponse(c, "Project ID is required")
		return
	}

	project, err := h.service.GetProject(userID, tenantID, projectID)
	if err != nil {
		if err.Error() == "access denied" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.NotFoundResponse(c, "Project not found")
		return
	}

	utils.SuccessResponse(c, project, "Project retrieved successfully")
}

// UpdateProject handles PUT /projects/:id
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		utils.BadRequestResponse(c, "Project ID is required")
		return
	}

	var req tenant_models.ProjectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	project, err := h.service.UpdateProject(userID, tenantID, projectID, &req)
	if err != nil {
		if err.Error() == "access denied" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update project")
		return
	}

	utils.SuccessResponse(c, project, "Project updated successfully")
}

// DeleteProject handles DELETE /projects/:id
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		utils.BadRequestResponse(c, "Project ID is required")
		return
	}

	err = h.service.DeleteProject(userID, tenantID, projectID)
	if err != nil {
		if err.Error() == "access denied" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to delete project")
		return
	}

	utils.SuccessResponse(c, nil, "Project deleted successfully")
}

// ListProjects handles GET /projects
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Cap at 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Parse filters
	filters := repositories.ProjectFilters{
		Status:    c.Query("status"),
		LeadID:    c.Query("lead_id"),
		CreatedBy: c.Query("created_by"),
		MemberID:  c.Query("member_id"),
		Search:    c.Query("search"),
	}

	// Parse date filters
	if dateFromStr := c.Query("created_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			filters.CreatedFrom = &dateFrom
		}
	}
	if dateToStr := c.Query("created_to"); dateToStr != "" {
		if dateTo, err := time.Parse("2006-01-02", dateToStr); err == nil {
			filters.CreatedTo = &dateTo
		}
	}

	projects, total, err := h.service.ListProjects(userID, tenantID, limit, offset, filters)
	if err != nil {
		h.logger.Error("Failed to list projects", zap.Error(err))
		utils.InternalServerErrorResponse(c, "Failed to list projects")
		return
	}

	// Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	pagination := utils.Pagination{
		Page:       (offset / limit) + 1,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	}

	utils.PaginatedSuccessResponse(c, projects, pagination, "Projects retrieved successfully")
}

// AddProjectMember handles POST /projects/:id/members
func (h *ProjectHandler) AddProjectMember(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		utils.BadRequestResponse(c, "Project ID is required")
		return
	}

	var req tenant_models.ProjectMemberCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	member, err := h.service.AddProjectMember(userID, tenantID, projectID, &req)
	if err != nil {
		if err.Error() == "access denied" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to add project member")
		return
	}

	utils.CreatedResponse(c, member, "Project member added successfully")
}

// RemoveProjectMember handles DELETE /projects/:id/members/:user_id
func (h *ProjectHandler) RemoveProjectMember(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	projectID := c.Param("id")
	memberUserID := c.Param("user_id")
	
	if projectID == "" || memberUserID == "" {
		utils.BadRequestResponse(c, "Project ID and User ID are required")
		return
	}

	err = h.service.RemoveProjectMember(userID, tenantID, projectID, memberUserID)
	if err != nil {
		if err.Error() == "access denied" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to remove project member")
		return
	}

	utils.SuccessResponse(c, nil, "Project member removed successfully")
}

// ListProjectMembers handles GET /projects/:id/members
func (h *ProjectHandler) ListProjectMembers(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		utils.BadRequestResponse(c, "Project ID is required")
		return
	}

	members, err := h.service.ListProjectMembers(userID, tenantID, projectID)
	if err != nil {
		if err.Error() == "access denied" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to list project members")
		return
	}

	utils.SuccessResponse(c, members, "Project members retrieved successfully")
}

// UpdateProjectMemberRole handles PATCH /projects/:id/members/:user_id/role
func (h *ProjectHandler) UpdateProjectMemberRole(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	projectID := c.Param("id")
	memberUserID := c.Param("user_id")
	
	if projectID == "" || memberUserID == "" {
		utils.BadRequestResponse(c, "Project ID and User ID are required")
		return
	}

	var req tenant_models.ProjectMemberUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	member, err := h.service.UpdateProjectMemberRole(userID, tenantID, projectID, memberUserID, &req)
	if err != nil {
		if err.Error() == "access denied" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update member role")
		return
	}

	utils.SuccessResponse(c, member, "Member role updated successfully")
}

// GetProjectStats handles GET /projects/:id/stats
func (h *ProjectHandler) GetProjectStats(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		utils.BadRequestResponse(c, "Project ID is required")
		return
	}

	// Parse date range
	var dateFrom, dateTo *time.Time
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if df, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			dateFrom = &df
		}
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dt, err := time.Parse("2006-01-02", dateToStr); err == nil {
			dateTo = &dt
		}
	}

	stats, err := h.service.GetProjectStats(userID, tenantID, projectID, dateFrom, dateTo)
	if err != nil {
		if err.Error() == "access denied" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get project stats")
		return
	}

	utils.SuccessResponse(c, stats, "Project statistics retrieved successfully")
}

// GetUserProjectStats handles GET /projects/stats/user
func (h *ProjectHandler) GetUserProjectStats(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	// Parse date range
	var dateFrom, dateTo *time.Time
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if df, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			dateFrom = &df
		}
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dt, err := time.Parse("2006-01-02", dateToStr); err == nil {
			dateTo = &dt
		}
	}

	stats, err := h.service.GetUserProjectStats(userID, tenantID, dateFrom, dateTo)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user project stats")
		return
	}

	utils.SuccessResponse(c, stats, "User project statistics retrieved successfully")
}