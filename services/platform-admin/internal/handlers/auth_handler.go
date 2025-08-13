package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zen/shared/pkg/utils"
	"github.com/zen/services/platform-admin/internal/services"
)

type PlatformAuthHandler struct {
	service *services.PlatformAuthService
	logger  *zap.Logger
}

func NewPlatformAuthHandler(service *services.PlatformAuthService, logger *zap.Logger) *PlatformAuthHandler {
	return &PlatformAuthHandler{
		service: service,
		logger:  logger,
	}
}

// Login handles POST /auth/login
func (h *PlatformAuthHandler) Login(c *gin.Context) {
	var req services.PlatformLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	authResponse, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Platform admin login failed", zap.Error(err), zap.String("email", req.Email))
		if err.Error() == "invalid credentials" {
			utils.UnauthorizedResponse(c, "Invalid email or password")
			return
		}
		if err.Error() == "account locked" {
			utils.UnauthorizedResponse(c, "Account is temporarily locked due to multiple failed attempts")
			return
		}
		if err.Error() == "account suspended" {
			utils.UnauthorizedResponse(c, "Account is suspended")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Login failed")
		return
	}

	utils.SuccessResponse(c, authResponse)
}

// RefreshToken handles POST /auth/refresh
func (h *PlatformAuthHandler) RefreshToken(c *gin.Context) {
	var req services.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	authResponse, err := h.service.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Platform admin token refresh failed", zap.Error(err))
		utils.UnauthorizedResponse(c, "Invalid refresh token")
		return
	}

	utils.SuccessResponse(c, authResponse)
}

// ListAdmins handles GET /admins
func (h *PlatformAuthHandler) ListAdmins(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	role := c.Query("role")
	status := c.Query("status")

	admins, total, err := h.service.ListAdmins(c.Request.Context(), &services.ListAdminsRequest{
		Page:   page,
		Limit:  limit,
		Search: search,
		Role:   role,
		Status: status,
	})

	if err != nil {
		h.logger.Error("Failed to list platform admins", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list administrators")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"admins": admins,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetAdmin handles GET /admins/:id
func (h *PlatformAuthHandler) GetAdmin(c *gin.Context) {
	adminID := c.Param("id")

	admin, err := h.service.GetAdmin(c.Request.Context(), adminID)
	if err != nil {
		h.logger.Error("Failed to get platform admin", zap.Error(err), zap.String("admin_id", adminID))
		if err.Error() == "admin not found" {
			utils.NotFoundResponse(c, "Administrator not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get administrator")
		return
	}

	utils.SuccessResponse(c, gin.H{"admin": admin})
}

// CreateAdmin handles POST /admins
func (h *PlatformAuthHandler) CreateAdmin(c *gin.Context) {
	var req services.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	admin, err := h.service.CreateAdmin(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create platform admin", zap.Error(err))
		if err.Error() == "email already exists" {
			utils.ErrorResponse(c, http.StatusConflict, "Email already exists")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create administrator")
		return
	}

	utils.CreatedResponse(c, gin.H{"admin": admin})
}

// UpdateAdmin handles PUT /admins/:id
func (h *PlatformAuthHandler) UpdateAdmin(c *gin.Context) {
	adminID := c.Param("id")

	var req services.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	admin, err := h.service.UpdateAdmin(c.Request.Context(), adminID, &req)
	if err != nil {
		h.logger.Error("Failed to update platform admin", zap.Error(err), zap.String("admin_id", adminID))
		if err.Error() == "admin not found" {
			utils.NotFoundResponse(c, "Administrator not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update administrator")
		return
	}

	utils.SuccessResponse(c, gin.H{"admin": admin})
}

// DeleteAdmin handles DELETE /admins/:id
func (h *PlatformAuthHandler) DeleteAdmin(c *gin.Context) {
	adminID := c.Param("id")

	err := h.service.DeleteAdmin(c.Request.Context(), adminID)
	if err != nil {
		h.logger.Error("Failed to delete platform admin", zap.Error(err), zap.String("admin_id", adminID))
		if err.Error() == "admin not found" {
			utils.NotFoundResponse(c, "Administrator not found")
			return
		}
		if err.Error() == "cannot delete super admin" {
			utils.ErrorResponse(c, http.StatusForbidden, "Cannot delete super administrator")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete administrator")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Administrator deleted successfully"})
}