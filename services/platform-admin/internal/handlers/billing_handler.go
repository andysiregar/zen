package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zen/shared/pkg/utils"
	"github.com/zen/services/platform-admin/internal/services"
)

type BillingHandler struct {
	service *services.BillingService
	logger  *zap.Logger
}

func NewBillingHandler(service *services.BillingService, logger *zap.Logger) *BillingHandler {
	return &BillingHandler{
		service: service,
		logger:  logger,
	}
}

// GetOrganizationBilling handles GET /billing/organizations/:org_id
func (h *BillingHandler) GetOrganizationBilling(c *gin.Context) {
	orgID := c.Param("org_id")

	billing, err := h.service.GetOrganizationBilling(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get organization billing", zap.Error(err), zap.String("organization_id", orgID))
		if err.Error() == "organization not found" {
			utils.NotFoundResponse(c, "Organization not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get organization billing")
		return
	}

	utils.SuccessResponse(c, gin.H{"billing": billing})
}

// GetTenantBilling handles GET /billing/tenants/:tenant_id
func (h *BillingHandler) GetTenantBilling(c *gin.Context) {
	tenantID := c.Param("tenant_id")

	billing, err := h.service.GetTenantBilling(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant billing", zap.Error(err), zap.String("tenant_id", tenantID))
		if err.Error() == "tenant not found" {
			utils.NotFoundResponse(c, "Tenant not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get tenant billing")
		return
	}

	utils.SuccessResponse(c, gin.H{"billing": billing})
}

// ListInvoices handles GET /billing/invoices
func (h *BillingHandler) ListInvoices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	organizationID := c.Query("organization_id")
	tenantID := c.Query("tenant_id")
	status := c.Query("status")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	invoices, total, err := h.service.ListInvoices(c.Request.Context(), &services.ListInvoicesRequest{
		Page:           page,
		Limit:          limit,
		OrganizationID: organizationID,
		TenantID:       tenantID,
		Status:         status,
		DateFrom:       dateFrom,
		DateTo:         dateTo,
		SortBy:         sortBy,
		SortOrder:      sortOrder,
	})

	if err != nil {
		h.logger.Error("Failed to list invoices", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list invoices")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"invoices": invoices,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetInvoice handles GET /billing/invoices/:id
func (h *BillingHandler) GetInvoice(c *gin.Context) {
	invoiceID := c.Param("id")

	invoice, err := h.service.GetInvoice(c.Request.Context(), invoiceID)
	if err != nil {
		h.logger.Error("Failed to get invoice", zap.Error(err), zap.String("invoice_id", invoiceID))
		if err.Error() == "invoice not found" {
			utils.NotFoundResponse(c, "Invoice not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get invoice")
		return
	}

	utils.SuccessResponse(c, gin.H{"invoice": invoice})
}

// GetBillingReports handles GET /billing/reports
func (h *BillingHandler) GetBillingReports(c *gin.Context) {
	reportType := c.DefaultQuery("type", "summary")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	organizationID := c.Query("organization_id")
	tenantID := c.Query("tenant_id")

	reports, err := h.service.GetBillingReports(c.Request.Context(), &services.BillingReportsRequest{
		Type:           reportType,
		DateFrom:       dateFrom,
		DateTo:         dateTo,
		OrganizationID: organizationID,
		TenantID:       tenantID,
	})

	if err != nil {
		h.logger.Error("Failed to get billing reports", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get billing reports")
		return
	}

	utils.SuccessResponse(c, gin.H{"reports": reports})
}

// GenerateInvoice handles POST /billing/invoices
func (h *BillingHandler) GenerateInvoice(c *gin.Context) {
	var req services.GenerateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	invoice, err := h.service.GenerateInvoice(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to generate invoice", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate invoice")
		return
	}

	utils.CreatedResponse(c, gin.H{"invoice": invoice})
}

// ProcessRefund handles POST /billing/refunds
func (h *BillingHandler) ProcessRefund(c *gin.Context) {
	var req services.ProcessRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	refund, err := h.service.ProcessRefund(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to process refund", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process refund")
		return
	}

	utils.CreatedResponse(c, gin.H{"refund": refund})
}

// UpdateOrganizationBilling handles PUT /billing/organizations/:org_id
func (h *BillingHandler) UpdateOrganizationBilling(c *gin.Context) {
	orgID := c.Param("org_id")

	var req services.UpdateOrganizationBillingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	billing, err := h.service.UpdateOrganizationBilling(c.Request.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("Failed to update organization billing", zap.Error(err), zap.String("organization_id", orgID))
		if err.Error() == "organization not found" {
			utils.NotFoundResponse(c, "Organization not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update organization billing")
		return
	}

	utils.SuccessResponse(c, gin.H{"billing": billing})
}