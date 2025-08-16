package services

import (
	"time"
	"github.com/google/uuid"
	
	"background-jobs-service/internal/models"
	"background-jobs-service/internal/repositories"
)

type JobService interface {
	// Job management
	CreateJob(tenantID string, request *models.CreateJobRequest) (*models.Job, error)
	GetJob(tenantID, jobID string) (*models.Job, error)
	UpdateJob(tenantID, jobID string, request *models.UpdateJobRequest) (*models.Job, error)
	DeleteJob(tenantID, jobID string) error
	ListJobs(tenantID string, filters models.JobFilters, page, limit int) (*models.JobListResponse, error)
	RetryJob(tenantID, jobID string) error
	CancelJob(tenantID, jobID string) error
	
	// Scheduled jobs
	CreateScheduledJob(tenantID string, request *models.CreateScheduledJobRequest) (*models.ScheduledJob, error)
	GetScheduledJob(tenantID, scheduledJobID string) (*models.ScheduledJob, error)
	UpdateScheduledJob(tenantID, scheduledJobID string, request *models.CreateScheduledJobRequest) (*models.ScheduledJob, error)
	DeleteScheduledJob(tenantID, scheduledJobID string) error
	ListScheduledJobs(tenantID string, page, limit int) (*models.ScheduledJobListResponse, error)
	
	// Metrics
	GetJobMetrics(tenantID string, jobType string, dateFrom, dateTo *time.Time) (*models.JobMetrics, error)
	GetTenantStats(tenantID string) (map[string]int64, error)
}

type jobService struct {
	repo repositories.JobRepository
}

func NewJobService(repo repositories.JobRepository) JobService {
	return &jobService{
		repo: repo,
	}
}

func (s *jobService) CreateJob(tenantID string, request *models.CreateJobRequest) (*models.Job, error) {
	now := time.Now()
	job := &models.Job{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		Type:         request.Type,
		Status:       models.JobStatusPending,
		Priority:     request.Priority,
		Payload:      request.Payload,
		MaxRetries:   request.MaxRetries,
		RunAt:        request.RunAt,
		Tags:         request.Tags,
		Dependencies: request.Dependencies,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Set default values
	if job.Priority == 0 {
		job.Priority = models.JobPriorityNormal
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}
	if job.Payload == nil {
		job.Payload = make(map[string]interface{})
	}

	// If job should run in the future, set status to queued
	if job.RunAt != nil && job.RunAt.After(now) {
		job.Status = models.JobStatusQueued
	}

	err := s.repo.CreateJob(tenantID, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *jobService) GetJob(tenantID, jobID string) (*models.Job, error) {
	return s.repo.GetJob(tenantID, jobID)
}

func (s *jobService) UpdateJob(tenantID, jobID string, request *models.UpdateJobRequest) (*models.Job, error) {
	job, err := s.repo.GetJob(tenantID, jobID)
	if err != nil {
		return nil, err
	}

	// Update only provided fields
	if request.Status != "" {
		job.Status = request.Status
	}
	if request.Priority > 0 {
		job.Priority = request.Priority
	}
	if request.RunAt != nil {
		job.RunAt = request.RunAt
	}
	job.UpdatedAt = time.Now()

	err = s.repo.UpdateJob(tenantID, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *jobService) DeleteJob(tenantID, jobID string) error {
	return s.repo.DeleteJob(tenantID, jobID)
}

func (s *jobService) ListJobs(tenantID string, filters models.JobFilters, page, limit int) (*models.JobListResponse, error) {
	offset := (page - 1) * limit
	jobs, total, err := s.repo.ListJobs(tenantID, filters, limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.JobListResponse{
		Jobs:       jobs,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (s *jobService) RetryJob(tenantID, jobID string) error {
	job, err := s.repo.GetJob(tenantID, jobID)
	if err != nil {
		return err
	}

	if job.Status != models.JobStatusFailed {
		return models.ErrJobInvalidState
	}

	// Reset job for retry
	job.Status = models.JobStatusPending
	job.StartedAt = nil
	job.CompletedAt = nil
	job.ErrorMessage = ""
	job.UpdatedAt = time.Now()

	return s.repo.UpdateJob(tenantID, job)
}

func (s *jobService) CancelJob(tenantID, jobID string) error {
	job, err := s.repo.GetJob(tenantID, jobID)
	if err != nil {
		return err
	}

	if job.Status == models.JobStatusCompleted || job.Status == models.JobStatusFailed {
		return models.ErrJobInvalidState
	}

	job.Status = models.JobStatusCancelled
	job.UpdatedAt = time.Now()

	return s.repo.UpdateJob(tenantID, job)
}

func (s *jobService) CreateScheduledJob(tenantID string, request *models.CreateScheduledJobRequest) (*models.ScheduledJob, error) {
	now := time.Now()
	scheduledJob := &models.ScheduledJob{
		ID:             uuid.New().String(),
		TenantID:       tenantID,
		Name:           request.Name,
		JobType:        request.JobType,
		CronExpression: request.CronExpression,
		Payload:        request.Payload,
		Priority:       request.Priority,
		MaxRetries:     request.MaxRetries,
		IsActive:       request.IsActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Set default values
	if scheduledJob.Priority == 0 {
		scheduledJob.Priority = models.JobPriorityNormal
	}
	if scheduledJob.MaxRetries == 0 {
		scheduledJob.MaxRetries = 3
	}
	if scheduledJob.Payload == nil {
		scheduledJob.Payload = make(map[string]interface{})
	}

	// TODO: Calculate next run time from cron expression
	// For now, set it to run in 1 hour
	nextRun := now.Add(time.Hour)
	scheduledJob.NextRunAt = &nextRun

	err := s.repo.CreateScheduledJob(tenantID, scheduledJob)
	if err != nil {
		return nil, err
	}

	return scheduledJob, nil
}

func (s *jobService) GetScheduledJob(tenantID, scheduledJobID string) (*models.ScheduledJob, error) {
	return s.repo.GetScheduledJob(tenantID, scheduledJobID)
}

func (s *jobService) UpdateScheduledJob(tenantID, scheduledJobID string, request *models.CreateScheduledJobRequest) (*models.ScheduledJob, error) {
	scheduledJob, err := s.repo.GetScheduledJob(tenantID, scheduledJobID)
	if err != nil {
		return nil, err
	}

	// Update fields
	scheduledJob.Name = request.Name
	scheduledJob.JobType = request.JobType
	scheduledJob.CronExpression = request.CronExpression
	scheduledJob.Payload = request.Payload
	scheduledJob.Priority = request.Priority
	scheduledJob.MaxRetries = request.MaxRetries
	scheduledJob.IsActive = request.IsActive
	scheduledJob.UpdatedAt = time.Now()

	// TODO: Recalculate next run time from new cron expression

	err = s.repo.UpdateScheduledJob(tenantID, scheduledJob)
	if err != nil {
		return nil, err
	}

	return scheduledJob, nil
}

func (s *jobService) DeleteScheduledJob(tenantID, scheduledJobID string) error {
	return s.repo.DeleteScheduledJob(tenantID, scheduledJobID)
}

func (s *jobService) ListScheduledJobs(tenantID string, page, limit int) (*models.ScheduledJobListResponse, error) {
	offset := (page - 1) * limit
	scheduledJobs, total, err := s.repo.ListScheduledJobs(tenantID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.ScheduledJobListResponse{
		ScheduledJobs: scheduledJobs,
		TotalCount:    total,
		Page:          page,
		Limit:         limit,
	}, nil
}

func (s *jobService) GetJobMetrics(tenantID string, jobType string, dateFrom, dateTo *time.Time) (*models.JobMetrics, error) {
	return s.repo.GetJobMetrics(tenantID, jobType, dateFrom, dateTo)
}

func (s *jobService) GetTenantStats(tenantID string) (map[string]int64, error) {
	return s.repo.GetTenantJobStats(tenantID)
}