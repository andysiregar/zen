package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"background-jobs-service/internal/models"
	"background-jobs-service/internal/services"
	"github.com/zen/shared/pkg/utils"
)

type JobHandler struct {
	jobService services.JobService
	logger     *zap.Logger
}

func NewJobHandler(jobService services.JobService, logger *zap.Logger) *JobHandler {
	return &JobHandler{
		jobService: jobService,
		logger:     logger,
	}
}

func (h *JobHandler) GetJobs(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	// Parse pagination
	page := 1
	limit := 20
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse filters
	filters := models.JobFilters{
		Status: models.JobStatus(c.Query("status")),
		Type:   c.Query("type"),
	}

	if priorityStr := c.Query("priority"); priorityStr != "" {
		if p, err := strconv.Atoi(priorityStr); err == nil {
			filters.Priority = models.JobPriority(p)
		}
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	jobs, err := h.jobService.ListJobs(tenantID, filters, page, limit)
	if err != nil {
		h.logger.Error("Failed to get jobs", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get jobs")
		return
	}

	utils.SuccessResponse(c, jobs)
}

func (h *JobHandler) CreateJob(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	var req models.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	job, err := h.jobService.CreateJob(tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to create job", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create job")
		return
	}

	utils.SuccessResponse(c, job)
}

func (h *JobHandler) GetJob(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	jobID := c.Param("id")
	if jobID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Job ID required")
		return
	}

	job, err := h.jobService.GetJob(tenantID, jobID)
	if err != nil {
		h.logger.Error("Failed to get job", zap.Error(err))
		utils.ErrorResponse(c, http.StatusNotFound, "Job not found")
		return
	}

	utils.SuccessResponse(c, job)
}

func (h *JobHandler) UpdateJob(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	jobID := c.Param("id")
	if jobID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Job ID required")
		return
	}

	var req models.UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	job, err := h.jobService.UpdateJob(tenantID, jobID, &req)
	if err != nil {
		h.logger.Error("Failed to update job", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update job")
		return
	}

	utils.SuccessResponse(c, job)
}

func (h *JobHandler) DeleteJob(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	jobID := c.Param("id")
	if jobID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Job ID required")
		return
	}

	err := h.jobService.DeleteJob(tenantID, jobID)
	if err != nil {
		h.logger.Error("Failed to delete job", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete job")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Job deleted successfully"})
}

func (h *JobHandler) RetryJob(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	jobID := c.Param("id")
	if jobID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Job ID required")
		return
	}

	err := h.jobService.RetryJob(tenantID, jobID)
	if err != nil {
		h.logger.Error("Failed to retry job", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retry job")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Job retry initiated successfully"})
}

func (h *JobHandler) CancelJob(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	jobID := c.Param("id")
	if jobID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Job ID required")
		return
	}

	err := h.jobService.CancelJob(tenantID, jobID)
	if err != nil {
		h.logger.Error("Failed to cancel job", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to cancel job")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Job cancelled successfully"})
}

func (h *JobHandler) GetJobMetrics(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	jobType := c.Query("type")
	
	var dateFrom, dateTo *time.Time
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if df, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			dateFrom = &df
		}
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dt, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			dateTo = &dt
		}
	}

	metrics, err := h.jobService.GetJobMetrics(tenantID, jobType, dateFrom, dateTo)
	if err != nil {
		h.logger.Error("Failed to get job metrics", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get job metrics")
		return
	}

	utils.SuccessResponse(c, metrics)
}

func (h *JobHandler) GetTenantStats(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	stats, err := h.jobService.GetTenantStats(tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant stats", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get tenant stats")
		return
	}

	utils.SuccessResponse(c, stats)
}

// Scheduled Jobs Handlers

func (h *JobHandler) GetScheduledJobs(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	// Parse pagination
	page := 1
	limit := 20
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	scheduledJobs, err := h.jobService.ListScheduledJobs(tenantID, page, limit)
	if err != nil {
		h.logger.Error("Failed to get scheduled jobs", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get scheduled jobs")
		return
	}

	utils.SuccessResponse(c, scheduledJobs)
}

func (h *JobHandler) CreateScheduledJob(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	var req models.CreateScheduledJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	scheduledJob, err := h.jobService.CreateScheduledJob(tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to create scheduled job", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create scheduled job")
		return
	}

	utils.SuccessResponse(c, scheduledJob)
}

func (h *JobHandler) GetScheduledJob(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	scheduledJobID := c.Param("id")
	if scheduledJobID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Scheduled Job ID required")
		return
	}

	scheduledJob, err := h.jobService.GetScheduledJob(tenantID, scheduledJobID)
	if err != nil {
		h.logger.Error("Failed to get scheduled job", zap.Error(err))
		utils.ErrorResponse(c, http.StatusNotFound, "Scheduled job not found")
		return
	}

	utils.SuccessResponse(c, scheduledJob)
}

func (h *JobHandler) UpdateScheduledJob(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	scheduledJobID := c.Param("id")
	if scheduledJobID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Scheduled Job ID required")
		return
	}

	var req models.CreateScheduledJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	scheduledJob, err := h.jobService.UpdateScheduledJob(tenantID, scheduledJobID, &req)
	if err != nil {
		h.logger.Error("Failed to update scheduled job", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update scheduled job")
		return
	}

	utils.SuccessResponse(c, scheduledJob)
}

func (h *JobHandler) DeleteScheduledJob(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tenant ID required")
		return
	}

	scheduledJobID := c.Param("id")
	if scheduledJobID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Scheduled Job ID required")
		return
	}

	err := h.jobService.DeleteScheduledJob(tenantID, scheduledJobID)
	if err != nil {
		h.logger.Error("Failed to delete scheduled job", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete scheduled job")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Scheduled job deleted successfully"})
}