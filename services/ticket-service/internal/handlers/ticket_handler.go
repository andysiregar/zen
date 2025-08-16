package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zen/shared/pkg/middleware"
	"github.com/zen/shared/pkg/models"
	"github.com/zen/shared/pkg/utils"
	"ticket-service/internal/repositories"
	"ticket-service/internal/services"
)

type TicketHandler struct {
	service services.TicketService
	logger  *zap.Logger
}

func NewTicketHandler(service services.TicketService, logger *zap.Logger) *TicketHandler {
	return &TicketHandler{
		service: service,
		logger:  logger,
	}
}

// Helper function to get user and tenant context
func (h *TicketHandler) getUserAndTenantContext(c *gin.Context) (string, string, error) {
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

// ListTickets handles GET /tickets
func (h *TicketHandler) ListTickets(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Parse filters
	filters := repositories.TicketFilters{
		Status:     c.Query("status"),
		Priority:   c.Query("priority"),
		Type:       c.Query("type"),
		AssigneeID: c.Query("assignee_id"),
		ReporterID: c.Query("reporter_id"),
		ProjectID:  c.Query("project_id"),
		Category:   c.Query("category"),
	}

	// Parse date filters
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			filters.DateFrom = &t
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.Parse(time.RFC3339, dateTo); err == nil {
			filters.DateTo = &t
		}
	}

	tickets, total, err := h.service.ListTickets(userID, tenantID, limit, offset, filters)
	if err != nil {
		h.logger.Error("Failed to list tickets", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list tickets")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"tickets": tickets,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// CreateTicket handles POST /tickets
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	var req models.TicketCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data: "+err.Error())
		return
	}

	ticket, err := h.service.CreateTicket(userID, tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to create ticket", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create ticket")
		return
	}

	utils.CreatedResponse(c, gin.H{"ticket": ticket})
}

// GetTicket handles GET /tickets/:id
func (h *TicketHandler) GetTicket(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	ticketID := c.Param("id")
	if ticketID == "" {
		utils.BadRequestResponse(c, "Ticket ID is required")
		return
	}

	ticket, err := h.service.GetTicket(userID, tenantID, ticketID)
	if err != nil {
		if err.Error() == "access denied: user cannot view this ticket" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		h.logger.Error("Failed to get ticket", zap.Error(err))
		utils.NotFoundResponse(c, "Ticket not found")
		return
	}

	utils.SuccessResponse(c, gin.H{"ticket": ticket})
}

// UpdateTicket handles PUT /tickets/:id
func (h *TicketHandler) UpdateTicket(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	ticketID := c.Param("id")
	if ticketID == "" {
		utils.BadRequestResponse(c, "Ticket ID is required")
		return
	}

	var req models.TicketUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data: "+err.Error())
		return
	}

	ticket, err := h.service.UpdateTicket(userID, tenantID, ticketID, &req)
	if err != nil {
		if err.Error() == "access denied: user cannot update this ticket" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		h.logger.Error("Failed to update ticket", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update ticket")
		return
	}

	utils.SuccessResponse(c, gin.H{"ticket": ticket})
}

// DeleteTicket handles DELETE /tickets/:id
func (h *TicketHandler) DeleteTicket(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	ticketID := c.Param("id")
	if ticketID == "" {
		utils.BadRequestResponse(c, "Ticket ID is required")
		return
	}

	err = h.service.DeleteTicket(userID, tenantID, ticketID)
	if err != nil {
		if err.Error() == "access denied: user cannot delete this ticket" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		h.logger.Error("Failed to delete ticket", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete ticket")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Ticket deleted successfully"})
}

// AssignTicket handles PATCH /tickets/:id/assign
func (h *TicketHandler) AssignTicket(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	ticketID := c.Param("id")
	if ticketID == "" {
		utils.BadRequestResponse(c, "Ticket ID is required")
		return
	}

	var req struct {
		AssigneeID string `json:"assignee_id" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data: "+err.Error())
		return
	}

	ticket, err := h.service.AssignTicket(userID, tenantID, ticketID, req.AssigneeID)
	if err != nil {
		h.logger.Error("Failed to assign ticket", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to assign ticket")
		return
	}

	utils.SuccessResponse(c, gin.H{"ticket": ticket})
}

// UpdateTicketStatus handles PATCH /tickets/:id/status
func (h *TicketHandler) UpdateTicketStatus(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	ticketID := c.Param("id")
	if ticketID == "" {
		utils.BadRequestResponse(c, "Ticket ID is required")
		return
	}

	var req struct {
		Status models.TicketStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data: "+err.Error())
		return
	}

	ticket, err := h.service.UpdateTicketStatus(userID, tenantID, ticketID, req.Status)
	if err != nil {
		h.logger.Error("Failed to update ticket status", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update ticket status")
		return
	}

	utils.SuccessResponse(c, gin.H{"ticket": ticket})
}

// UpdateTicketPriority handles PATCH /tickets/:id/priority
func (h *TicketHandler) UpdateTicketPriority(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	ticketID := c.Param("id")
	if ticketID == "" {
		utils.BadRequestResponse(c, "Ticket ID is required")
		return
	}

	var req struct {
		Priority models.TicketPriority `json:"priority" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data: "+err.Error())
		return
	}

	ticket, err := h.service.UpdateTicketPriority(userID, tenantID, ticketID, req.Priority)
	if err != nil {
		h.logger.Error("Failed to update ticket priority", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update ticket priority")
		return
	}

	utils.SuccessResponse(c, gin.H{"ticket": ticket})
}

// GetTicketComments handles GET /tickets/:id/comments
func (h *TicketHandler) GetTicketComments(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	ticketID := c.Param("id")
	if ticketID == "" {
		utils.BadRequestResponse(c, "Ticket ID is required")
		return
	}

	includeInternal := c.Query("include_internal") == "true"

	comments, err := h.service.GetComments(userID, tenantID, ticketID, includeInternal)
	if err != nil {
		h.logger.Error("Failed to get ticket comments", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get comments")
		return
	}

	utils.SuccessResponse(c, gin.H{"comments": comments})
}

// CreateTicketComment handles POST /tickets/:id/comments
func (h *TicketHandler) CreateTicketComment(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	ticketID := c.Param("id")
	if ticketID == "" {
		utils.BadRequestResponse(c, "Ticket ID is required")
		return
	}

	var req models.TicketCommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data: "+err.Error())
		return
	}

	comment, err := h.service.CreateComment(userID, tenantID, ticketID, &req)
	if err != nil {
		h.logger.Error("Failed to create comment", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	utils.CreatedResponse(c, gin.H{"comment": comment})
}

// UpdateComment handles PUT /comments/:comment_id
func (h *TicketHandler) UpdateComment(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	commentID := c.Param("comment_id")
	if commentID == "" {
		utils.BadRequestResponse(c, "Comment ID is required")
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data: "+err.Error())
		return
	}

	comment, err := h.service.UpdateComment(userID, tenantID, commentID, req.Content)
	if err != nil {
		h.logger.Error("Failed to update comment", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update comment")
		return
	}

	utils.SuccessResponse(c, gin.H{"comment": comment})
}

// DeleteComment handles DELETE /comments/:comment_id
func (h *TicketHandler) DeleteComment(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	commentID := c.Param("comment_id")
	if commentID == "" {
		utils.BadRequestResponse(c, "Comment ID is required")
		return
	}

	err = h.service.DeleteComment(userID, tenantID, commentID)
	if err != nil {
		h.logger.Error("Failed to delete comment", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Comment deleted successfully"})
}

// SearchTickets handles GET /tickets/search
func (h *TicketHandler) SearchTickets(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	tickets, total, err := h.service.SearchTickets(userID, tenantID, query, limit, offset)
	if err != nil {
		h.logger.Error("Failed to search tickets", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to search tickets")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"tickets": tickets,
		"query":   query,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetTicketStats handles GET /tickets/stats
func (h *TicketHandler) GetTicketStats(c *gin.Context) {
	userID, tenantID, err := h.getUserAndTenantContext(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	// Parse date range
	var dateFrom, dateTo *time.Time
	if from := c.Query("date_from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			dateFrom = &t
		}
	}
	if to := c.Query("date_to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			dateTo = &t
		}
	}

	stats, err := h.service.GetTicketStats(userID, tenantID, dateFrom, dateTo)
	if err != nil {
		h.logger.Error("Failed to get ticket stats", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get ticket statistics")
		return
	}

	utils.SuccessResponse(c, gin.H{"stats": stats})
}

// Placeholder handlers for attachment functionality
func (h *TicketHandler) GetTicketAttachments(c *gin.Context) {
	utils.SuccessResponse(c, gin.H{"attachments": []string{}})
}

func (h *TicketHandler) UploadAttachment(c *gin.Context) {
	utils.ErrorResponse(c, http.StatusNotImplemented, "Attachment upload not implemented yet")
}

func (h *TicketHandler) DeleteAttachment(c *gin.Context) {
	utils.ErrorResponse(c, http.StatusNotImplemented, "Attachment deletion not implemented yet")
}