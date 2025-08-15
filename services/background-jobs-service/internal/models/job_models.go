package models

import (
	"errors"
	"time"
)

var (
	ErrJobNotFound     = errors.New("job not found")
	ErrJobInvalidState = errors.New("job in invalid state")
	ErrJobTimeout      = errors.New("job timeout")
)

// JobStatus represents the current status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusQueued     JobStatus = "queued"
	JobStatusRunning    JobStatus = "running"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusRetrying   JobStatus = "retrying"
	JobStatusCancelled  JobStatus = "cancelled"
)

// JobPriority represents the priority level of a job
type JobPriority int

const (
	JobPriorityLow    JobPriority = 1
	JobPriorityNormal JobPriority = 5
	JobPriorityHigh   JobPriority = 10
	JobPriorityCritical JobPriority = 15
)

// Job represents a background job
type Job struct {
	ID            string                 `json:"id" gorm:"primaryKey"`
	TenantID      string                 `json:"tenant_id" gorm:"index"`
	Type          string                 `json:"type" gorm:"index"`
	Status        JobStatus              `json:"status" gorm:"index"`
	Priority      JobPriority            `json:"priority" gorm:"index"`
	Payload       map[string]interface{} `json:"payload" gorm:"type:jsonb"`
	Result        map[string]interface{} `json:"result,omitempty" gorm:"type:jsonb"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	RetryCount    int                    `json:"retry_count" gorm:"default:0"`
	MaxRetries    int                    `json:"max_retries" gorm:"default:3"`
	RunAt         *time.Time             `json:"run_at,omitempty" gorm:"index"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	LastError     *time.Time             `json:"last_error,omitempty"`
	TimeoutAt     *time.Time             `json:"timeout_at,omitempty"`
	CreatedBy     string                 `json:"created_by,omitempty"`
	Tags          []string               `json:"tags,omitempty" gorm:"type:text[]"`
	Dependencies  []string               `json:"dependencies,omitempty" gorm:"type:text[]"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ScheduledJob represents a recurring scheduled job
type ScheduledJob struct {
	ID            string                 `json:"id" gorm:"primaryKey"`
	TenantID      string                 `json:"tenant_id" gorm:"index"`
	Name          string                 `json:"name"`
	JobType       string                 `json:"job_type"`
	CronExpression string                `json:"cron_expression"`
	Payload       map[string]interface{} `json:"payload" gorm:"type:jsonb"`
	Priority      JobPriority            `json:"priority"`
	MaxRetries    int                    `json:"max_retries"`
	IsActive      bool                   `json:"is_active" gorm:"default:true"`
	NextRunAt     *time.Time             `json:"next_run_at" gorm:"index"`
	LastRunAt     *time.Time             `json:"last_run_at,omitempty"`
	LastJobID     string                 `json:"last_job_id,omitempty"`
	CreatedBy     string                 `json:"created_by,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// JobMetrics represents job execution metrics
type JobMetrics struct {
	TenantID          string    `json:"tenant_id"`
	JobType           string    `json:"job_type"`
	TotalJobs         int64     `json:"total_jobs"`
	CompletedJobs     int64     `json:"completed_jobs"`
	FailedJobs        int64     `json:"failed_jobs"`
	RunningJobs       int64     `json:"running_jobs"`
	QueuedJobs        int64     `json:"queued_jobs"`
	AvgExecutionTime  float64   `json:"avg_execution_time_seconds"`
	TotalExecutionTime float64  `json:"total_execution_time_seconds"`
	LastExecuted      *time.Time `json:"last_executed,omitempty"`
	SuccessRate       float64   `json:"success_rate"`
}

// WorkerStatus represents the status of a background worker
type WorkerStatus struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"` // idle, busy, stopped
	CurrentJob  *Job       `json:"current_job,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	LastPing    time.Time  `json:"last_ping"`
	JobsHandled int64      `json:"jobs_handled"`
}

// API Request/Response Models

type CreateJobRequest struct {
	Type         string                 `json:"type" binding:"required"`
	Payload      map[string]interface{} `json:"payload"`
	Priority     JobPriority            `json:"priority"`
	MaxRetries   int                    `json:"max_retries"`
	RunAt        *time.Time             `json:"run_at,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
}

type UpdateJobRequest struct {
	Status    JobStatus `json:"status,omitempty"`
	Priority  JobPriority `json:"priority,omitempty"`
	RunAt     *time.Time `json:"run_at,omitempty"`
}

type CreateScheduledJobRequest struct {
	Name           string                 `json:"name" binding:"required"`
	JobType        string                 `json:"job_type" binding:"required"`
	CronExpression string                 `json:"cron_expression" binding:"required"`
	Payload        map[string]interface{} `json:"payload"`
	Priority       JobPriority            `json:"priority"`
	MaxRetries     int                    `json:"max_retries"`
	IsActive       bool                   `json:"is_active"`
}

type JobListResponse struct {
	Jobs       []*Job `json:"jobs"`
	TotalCount int64  `json:"total_count"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

type ScheduledJobListResponse struct {
	ScheduledJobs []*ScheduledJob `json:"scheduled_jobs"`
	TotalCount    int64           `json:"total_count"`
	Page          int             `json:"page"`
	Limit         int             `json:"limit"`
}

type JobFilters struct {
	Status   JobStatus
	Type     string
	Priority JobPriority
	Tags     []string
	DateFrom *time.Time
	DateTo   *time.Time
}

// Job execution context
type JobContext struct {
	Job       *Job
	TenantID  string
	Cancel    chan bool
	Progress  chan int
	Log       func(message string, fields ...interface{})
	UpdateJob func(updates map[string]interface{}) error
}