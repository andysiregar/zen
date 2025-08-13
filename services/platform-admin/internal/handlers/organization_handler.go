package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zen/shared/pkg/utils"
	"github.com/zen/services/platform-admin/internal/services"
)

type OrganizationHandler struct {
	service *services.OrganizationService
	logger  *zap.Logger
}

func NewOrganizationHandler(service *services.OrganizationService, logger *zap.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		service: service,
		logger:  logger,
	}
}

// ListOrganizations handles GET /organizations
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	status := c.Query("status")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	organizations, total, err := h.service.ListOrganizations(c.Request.Context(), &services.ListOrganizationsRequest{
		Page:      page,
		Limit:     limit,
		Search:    search,
		Status:    status,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	})

	if err != nil {
		h.logger.Error("Failed to list organizations", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list organizations")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"organizations": organizations,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetOrganization handles GET /organizations/:id
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	orgID := c.Param("id")

	organization, err := h.service.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get organization", zap.Error(err), zap.String("organization_id", orgID))
		if err.Error() == "organization not found" {
			utils.NotFoundResponse(c, "Organization not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get organization")
		return
	}

	utils.SuccessResponse(c, gin.H{"organization": organization})
}

// GetOrganizationTenants handles GET /organizations/:id/tenants
func (h *OrganizationHandler) GetOrganizationTenants(c *gin.Context) {
	orgID := c.Param("id")

	tenants, err := h.service.GetOrganizationTenants(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get organization tenants", zap.Error(err), zap.String("organization_id", orgID))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get organization tenants")
		return
	}

	utils.SuccessResponse(c, gin.H{"tenants": tenants})
}

// GetOrganizationUsers handles GET /organizations/:id/users
func (h *OrganizationHandler) GetOrganizationUsers(c *gin.Context) {
	orgID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	users, total, err := h.service.GetOrganizationUsers(c.Request.Context(), orgID, page, limit)
	if err != nil {
		h.logger.Error("Failed to get organization users", zap.Error(err), zap.String("organization_id", orgID))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get organization users")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetOrganizationStats handles GET /organizations/:id/stats
func (h *OrganizationHandler) GetOrganizationStats(c *gin.Context) {
	orgID := c.Param("id")

	stats, err := h.service.GetOrganizationStats(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get organization stats", zap.Error(err), zap.String("organization_id", orgID))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get organization stats")
		return
	}

	utils.SuccessResponse(c, gin.H{"stats": stats})
}

// CreateOrganization handles POST /organizations
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req services.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	organization, err := h.service.CreateOrganization(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create organization", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create organization")
		return
	}

	utils.CreatedResponse(c, gin.H{"organization": organization})
}

// UpdateOrganization handles PUT /organizations/:id
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	orgID := c.Param("id")

	var req services.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	organization, err := h.service.UpdateOrganization(c.Request.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("Failed to update organization", zap.Error(err), zap.String("organization_id", orgID))
		if err.Error() == "organization not found" {
			utils.NotFoundResponse(c, "Organization not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update organization")
		return
	}

	utils.SuccessResponse(c, gin.H{"organization": organization})
}

// SuspendOrganization handles PATCH /organizations/:id/suspend
func (h *OrganizationHandler) SuspendOrganization(c *gin.Context) {
	orgID := c.Param("id")

	var req services.SuspendOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err := h.service.SuspendOrganization(c.Request.Context(), orgID, req.Reason)
	if err != nil {
		h.logger.Error("Failed to suspend organization", zap.Error(err), zap.String("organization_id", orgID))
		if err.Error() == "organization not found" {
			utils.NotFoundResponse(c, "Organization not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to suspend organization")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Organization suspended successfully"})
}

// ActivateOrganization handles PATCH /organizations/:id/activate
func (h *OrganizationHandler) ActivateOrganization(c *gin.Context) {
	orgID := c.Param("id")

	err := h.service.ActivateOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to activate organization", zap.Error(err), zap.String("organization_id", orgID))
		if err.Error() == "organization not found" {
			utils.NotFoundResponse(c, "Organization not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to activate organization")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Organization activated successfully"})
}

// DeleteOrganization handles DELETE /organizations/:id
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	orgID := c.Param("id")

	var req services.DeleteOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body provided, default to soft delete
		req.HardDelete = false
	}

	err := h.service.DeleteOrganization(c.Request.Context(), orgID, req.HardDelete)
	if err != nil {
		h.logger.Error("Failed to delete organization", zap.Error(err), zap.String("organization_id", orgID))
		if err.Error() == "organization not found" {
			utils.NotFoundResponse(c, "Organization not found")
			return
		}
		if err.Error() == "organization has active tenants" {
			utils.ErrorResponse(c, http.StatusConflict, "Cannot delete organization with active tenants")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete organization")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Organization deleted successfully"})
}