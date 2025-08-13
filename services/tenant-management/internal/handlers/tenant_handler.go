package handlers

import (
	"strconv"
	
	"github.com/gin-gonic/gin"
	"github.com/zen/shared/pkg/models"
	"tenant-management/internal/services"
	"github.com/zen/shared/pkg/utils"
)

type TenantHandler struct {
	service services.TenantService
}

func NewTenantHandler(service services.TenantService) *TenantHandler {
	return &TenantHandler{service: service}
}

func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req models.TenantCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}
	
	tenant, err := h.service.CreateTenant(req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create tenant: "+err.Error())
		return
	}
	
	utils.CreatedResponse(c, tenant, "Tenant created successfully")
}

func (h *TenantHandler) GetTenant(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Tenant ID is required")
		return
	}
	
	tenant, err := h.service.GetTenant(id)
	if err != nil {
		utils.NotFoundResponse(c, "Tenant not found: "+err.Error())
		return
	}
	
	utils.SuccessResponse(c, tenant, "Tenant retrieved successfully")
}

func (h *TenantHandler) GetTenantBySlug(c *gin.Context) {
	slug := c.Query("slug")
	if slug == "" {
		utils.BadRequestResponse(c, "Slug query parameter is required")
		return
	}
	
	tenant, err := h.service.GetTenantBySlug(slug)
	if err != nil {
		utils.NotFoundResponse(c, "Tenant not found: "+err.Error())
		return
	}
	
	utils.SuccessResponse(c, tenant, "Tenant retrieved successfully")
}

func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Tenant ID is required")
		return
	}
	
	var req models.TenantUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}
	
	tenant, err := h.service.UpdateTenant(id, req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update tenant: "+err.Error())
		return
	}
	
	utils.SuccessResponse(c, tenant, "Tenant updated successfully")
}

func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequestResponse(c, "Tenant ID is required")
		return
	}
	
	err := h.service.DeleteTenant(id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete tenant: "+err.Error())
		return
	}
	
	utils.SuccessResponse(c, gin.H{"message": "Tenant deleted successfully"}, "Success")
}

func (h *TenantHandler) ListTenants(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid limit parameter: "+err.Error())
		return
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid offset parameter: "+err.Error())
		return
	}
	
	tenants, total, err := h.service.ListTenants(limit, offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to list tenants: "+err.Error())
		return
	}
	
	response := gin.H{
		"tenants": tenants,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	}
	
	utils.SuccessResponse(c, response, "Tenants retrieved successfully")
}