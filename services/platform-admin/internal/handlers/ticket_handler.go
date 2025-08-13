package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zen/shared/pkg/utils"
	"github.com/zen/services/platform-admin/internal/services"
)

type CrossTenantTicketHandler struct {
	service *services.CrossTenantTicketService
	logger  *zap.Logger
}

func NewCrossTenantTicketHandler(service *services.CrossTenantTicketService, logger *zap.Logger) *CrossTenantTicketHandler {
	return &CrossTenantTicketHandler{
		service: service,
		logger:  logger,
	}
}

// ListAllTickets handles GET /tickets - lists tickets across all tenants
func (h *CrossTenantTicketHandler) ListAllTickets(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	priority := c.Query("priority")
	tenantID := c.Query("tenant_id")
	organizationID := c.Query("organization_id")
	assignedTo := c.Query("assigned_to")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	tickets, total, err := h.service.ListAllTickets(c.Request.Context(), &services.ListTicketsRequest{
		Page:           page,
		Limit:          limit,
		Status:         status,
		Priority:       priority,
		TenantID:       tenantID,
		OrganizationID: organizationID,
		AssignedTo:     assignedTo,
		SortBy:         sortBy,
		SortOrder:      sortOrder,
	})

	if err != nil {
		h.logger.Error("Failed to list all tickets", zap.Error(err))
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

// SearchTickets handles GET /tickets/search - search tickets across all tenants
func (h *CrossTenantTicketHandler) SearchTickets(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	tenantID := c.Query("tenant_id")
	organizationID := c.Query("organization_id")

	tickets, total, err := h.service.SearchTickets(c.Request.Context(), &services.SearchTicketsRequest{
		Query:          query,
		Page:           page,
		Limit:          limit,
		TenantID:       tenantID,
		OrganizationID: organizationID,
	})

	if err != nil {
		h.logger.Error("Failed to search tickets", zap.Error(err), zap.String("query", query))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to search tickets")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"tickets": tickets,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
		"query": query,
	})
}

// GetTicketStats handles GET /tickets/stats - get ticket statistics across all tenants
func (h *CrossTenantTicketHandler) GetTicketStats(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	organizationID := c.Query("organization_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	stats, err := h.service.GetTicketStats(c.Request.Context(), &services.TicketStatsRequest{
		TenantID:       tenantID,
		OrganizationID: organizationID,
		DateFrom:       dateFrom,
		DateTo:         dateTo,
	})

	if err != nil {
		h.logger.Error("Failed to get ticket stats", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get ticket statistics")
		return
	}

	utils.SuccessResponse(c, gin.H{"stats": stats})
}

// GetTicket handles GET /tickets/:id - get specific ticket with tenant context
func (h *CrossTenantTicketHandler) GetTicket(c *gin.Context) {
	ticketID := c.Param("id")

	ticket, err := h.service.GetTicket(c.Request.Context(), ticketID)
	if err != nil {
		h.logger.Error("Failed to get ticket", zap.Error(err), zap.String("ticket_id", ticketID))
		if err.Error() == "ticket not found" {
			utils.NotFoundResponse(c, "Ticket not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get ticket")
		return
	}

	utils.SuccessResponse(c, gin.H{"ticket": ticket})
}

// GetTenantTickets handles GET /tickets/tenant/:tenant_id - get tickets for specific tenant
func (h *CrossTenantTicketHandler) GetTenantTickets(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	priority := c.Query("priority")

	tickets, total, err := h.service.GetTenantTickets(c.Request.Context(), tenantID, &services.TenantTicketsRequest{
		Page:     page,
		Limit:    limit,
		Status:   status,
		Priority: priority,
	})

	if err != nil {
		h.logger.Error("Failed to get tenant tickets", zap.Error(err), zap.String("tenant_id", tenantID))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get tenant tickets")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"tickets": tickets,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
		"tenant_id": tenantID,
	})
}

// GetOrganizationTickets handles GET /tickets/organization/:org_id - get tickets for specific organization
func (h *CrossTenantTicketHandler) GetOrganizationTickets(c *gin.Context) {
	orgID := c.Param("org_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	priority := c.Query("priority")

	tickets, total, err := h.service.GetOrganizationTickets(c.Request.Context(), orgID, &services.OrganizationTicketsRequest{
		Page:     page,
		Limit:    limit,
		Status:   status,
		Priority: priority,
	})

	if err != nil {
		h.logger.Error("Failed to get organization tickets", zap.Error(err), zap.String("organization_id", orgID))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get organization tickets")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"tickets": tickets,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
		"organization_id": orgID,
	})
}

// UpdateTicket handles PUT /tickets/:id - update ticket status, priority, etc.
func (h *CrossTenantTicketHandler) UpdateTicket(c *gin.Context) {
	ticketID := c.Param("id")

	var req services.UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ticket, err := h.service.UpdateTicket(c.Request.Context(), ticketID, &req)
	if err != nil {
		h.logger.Error("Failed to update ticket", zap.Error(err), zap.String("ticket_id", ticketID))
		if err.Error() == "ticket not found" {
			utils.NotFoundResponse(c, "Ticket not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update ticket")
		return
	}

	utils.SuccessResponse(c, gin.H{"ticket": ticket})
}

// AssignTicket handles PATCH /tickets/:id/assign - assign ticket to an agent
func (h *CrossTenantTicketHandler) AssignTicket(c *gin.Context) {
	ticketID := c.Param("id")

	var req services.AssignTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ticket, err := h.service.AssignTicket(c.Request.Context(), ticketID, &req)
	if err != nil {
		h.logger.Error("Failed to assign ticket", zap.Error(err), zap.String("ticket_id", ticketID))
		if err.Error() == "ticket not found" {
			utils.NotFoundResponse(c, "Ticket not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to assign ticket")
		return
	}

	utils.SuccessResponse(c, gin.H{"ticket": ticket})
}

// CloseTicket handles PATCH /tickets/:id/close - close a ticket
func (h *CrossTenantTicketHandler) CloseTicket(c *gin.Context) {
	ticketID := c.Param("id")

	var req services.CloseTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ticket, err := h.service.CloseTicket(c.Request.Context(), ticketID, &req)
	if err != nil {
		h.logger.Error("Failed to close ticket", zap.Error(err), zap.String("ticket_id", ticketID))
		if err.Error() == "ticket not found" {
			utils.NotFoundResponse(c, "Ticket not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to close ticket")
		return
	}

	utils.SuccessResponse(c, gin.H{"ticket": ticket})
}

// DeleteTicket handles DELETE /tickets/:id - delete a ticket (usually soft delete)
func (h *CrossTenantTicketHandler) DeleteTicket(c *gin.Context) {
	ticketID := c.Param("id")

	var req services.DeleteTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body provided, default to soft delete
		req.HardDelete = false
	}

	err := h.service.DeleteTicket(c.Request.Context(), ticketID, req.HardDelete)
	if err != nil {
		h.logger.Error("Failed to delete ticket", zap.Error(err), zap.String("ticket_id", ticketID))
		if err.Error() == "ticket not found" {
			utils.NotFoundResponse(c, "Ticket not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete ticket")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Ticket deleted successfully"})
}