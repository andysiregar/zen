package repositories

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/zen/shared/pkg/database"
	"background-jobs-service/internal/models"
)

type JobRepository interface {
	// Job CRUD operations
	CreateJob(tenantID string, job *models.Job) error
	GetJob(tenantID, jobID string) (*models.Job, error)
	UpdateJob(tenantID string, job *models.Job) error
	DeleteJob(tenantID, jobID string) error
	ListJobs(tenantID string, filters models.JobFilters, limit, offset int) ([]*models.Job, int64, error)
	
	// Job queue operations
	GetNextJob(workerID string) (*models.Job, error)
	ClaimJob(jobID, workerID string) error
	ReleaseJob(jobID string) error
	GetJobsByStatus(tenantID string, status models.JobStatus) ([]*models.Job, error)
	GetReadyJobs(limit int) ([]*models.Job, error)
	
	// Scheduled job operations
	CreateScheduledJob(tenantID string, scheduledJob *models.ScheduledJob) error
	GetScheduledJob(tenantID, scheduledJobID string) (*models.ScheduledJob, error)
	UpdateScheduledJob(tenantID string, scheduledJob *models.ScheduledJob) error
	DeleteScheduledJob(tenantID, scheduledJobID string) error
	ListScheduledJobs(tenantID string, limit, offset int) ([]*models.ScheduledJob, int64, error)
	GetDueScheduledJobs() ([]*models.ScheduledJob, error)
	
	// Metrics and monitoring
	GetJobMetrics(tenantID string, jobType string, dateFrom, dateTo *time.Time) (*models.JobMetrics, error)
	GetTenantJobStats(tenantID string) (map[string]int64, error)
	CleanupOldJobs(olderThan time.Time) error
}

type jobRepository struct {
	tenantDBManager *database.TenantDatabaseManager
	masterDB        *database.DatabaseManager
}

func NewJobRepository(tenantDBManager *database.TenantDatabaseManager, masterDB *database.DatabaseManager) JobRepository {
	return &jobRepository{
		tenantDBManager: tenantDBManager,
		masterDB:        masterDB,
	}
}

func (r *jobRepository) CreateJob(tenantID string, job *models.Job) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	payloadJSON, _ := json.Marshal(job.Payload)
	resultJSON, _ := json.Marshal(job.Result)
	tagsJSON, _ := json.Marshal(job.Tags)
	dependenciesJSON, _ := json.Marshal(job.Dependencies)

	result := db.Exec(`
		INSERT INTO jobs (id, tenant_id, type, status, priority, payload, result, error_message, retry_count, max_retries, run_at, started_at, completed_at, last_error, timeout_at, created_by, tags, dependencies, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, job.ID, job.TenantID, job.Type, job.Status, job.Priority, string(payloadJSON), string(resultJSON),
		job.ErrorMessage, job.RetryCount, job.MaxRetries, job.RunAt, job.StartedAt, job.CompletedAt,
		job.LastError, job.TimeoutAt, job.CreatedBy, string(tagsJSON), string(dependenciesJSON),
		job.CreatedAt, job.UpdatedAt)

	return result.Error
}

func (r *jobRepository) GetJob(tenantID, jobID string) (*models.Job, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var job models.Job
	var payloadJSON, resultJSON, tagsJSON, dependenciesJSON string
	
	err = db.Raw(`
		SELECT id, tenant_id, type, status, priority, payload, result, error_message, retry_count, max_retries, run_at, started_at, completed_at, last_error, timeout_at, created_by, tags, dependencies, created_at, updated_at
		FROM jobs 
		WHERE id = ? AND tenant_id = ?
	`, jobID, tenantID).Row().Scan(
		&job.ID, &job.TenantID, &job.Type, &job.Status, &job.Priority, &payloadJSON, &resultJSON,
		&job.ErrorMessage, &job.RetryCount, &job.MaxRetries, &job.RunAt, &job.StartedAt, &job.CompletedAt,
		&job.LastError, &job.TimeoutAt, &job.CreatedBy, &tagsJSON, &dependenciesJSON,
		&job.CreatedAt, &job.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(payloadJSON), &job.Payload)
	json.Unmarshal([]byte(resultJSON), &job.Result)
	json.Unmarshal([]byte(tagsJSON), &job.Tags)
	json.Unmarshal([]byte(dependenciesJSON), &job.Dependencies)

	return &job, nil
}

func (r *jobRepository) UpdateJob(tenantID string, job *models.Job) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	payloadJSON, _ := json.Marshal(job.Payload)
	resultJSON, _ := json.Marshal(job.Result)
	tagsJSON, _ := json.Marshal(job.Tags)
	dependenciesJSON, _ := json.Marshal(job.Dependencies)

	result := db.Exec(`
		UPDATE jobs 
		SET type = ?, status = ?, priority = ?, payload = ?, result = ?, error_message = ?, retry_count = ?, max_retries = ?, run_at = ?, started_at = ?, completed_at = ?, last_error = ?, timeout_at = ?, created_by = ?, tags = ?, dependencies = ?, updated_at = ?
		WHERE id = ? AND tenant_id = ?
	`, job.Type, job.Status, job.Priority, string(payloadJSON), string(resultJSON), job.ErrorMessage,
		job.RetryCount, job.MaxRetries, job.RunAt, job.StartedAt, job.CompletedAt, job.LastError,
		job.TimeoutAt, job.CreatedBy, string(tagsJSON), string(dependenciesJSON), job.UpdatedAt,
		job.ID, tenantID)

	return result.Error
}

func (r *jobRepository) DeleteJob(tenantID, jobID string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		DELETE FROM jobs 
		WHERE id = ? AND tenant_id = ?
	`, jobID, tenantID)

	return result.Error
}

func (r *jobRepository) ListJobs(tenantID string, filters models.JobFilters, limit, offset int) ([]*models.Job, int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, 0, err
	}

	whereClause := "WHERE tenant_id = ?"
	args := []interface{}{tenantID}

	if filters.Status != "" {
		whereClause += " AND status = ?"
		args = append(args, filters.Status)
	}
	if filters.Type != "" {
		whereClause += " AND type = ?"
		args = append(args, filters.Type)
	}
	if filters.Priority > 0 {
		whereClause += " AND priority = ?"
		args = append(args, filters.Priority)
	}
	if filters.DateFrom != nil {
		whereClause += " AND created_at >= ?"
		args = append(args, filters.DateFrom)
	}
	if filters.DateTo != nil {
		whereClause += " AND created_at <= ?"
		args = append(args, filters.DateTo)
	}

	// Get total count
	var count int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM jobs %s", whereClause)
	db.Raw(countQuery, args...).Scan(&count)

	// Get jobs with pagination
	query := fmt.Sprintf(`
		SELECT id, tenant_id, type, status, priority, payload, result, error_message, retry_count, max_retries, run_at, started_at, completed_at, last_error, timeout_at, created_by, tags, dependencies, created_at, updated_at
		FROM jobs %s
		ORDER BY priority DESC, created_at ASC
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, limit, offset)
	rows, err := db.Raw(query, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var job models.Job
		var payloadJSON, resultJSON, tagsJSON, dependenciesJSON string
		
		rows.Scan(
			&job.ID, &job.TenantID, &job.Type, &job.Status, &job.Priority, &payloadJSON, &resultJSON,
			&job.ErrorMessage, &job.RetryCount, &job.MaxRetries, &job.RunAt, &job.StartedAt, &job.CompletedAt,
			&job.LastError, &job.TimeoutAt, &job.CreatedBy, &tagsJSON, &dependenciesJSON,
			&job.CreatedAt, &job.UpdatedAt,
		)

		json.Unmarshal([]byte(payloadJSON), &job.Payload)
		json.Unmarshal([]byte(resultJSON), &job.Result)
		json.Unmarshal([]byte(tagsJSON), &job.Tags)
		json.Unmarshal([]byte(dependenciesJSON), &job.Dependencies)

		jobs = append(jobs, &job)
	}

	return jobs, count, nil
}

func (r *jobRepository) GetNextJob(workerID string) (*models.Job, error) {
	// This would typically use Redis for distributed job queuing
	// For now, we'll use a simple database approach
	masterDB := r.masterDB.GetMasterDB()
	
	// Find the next available job across all tenants
	var job models.Job
	err := masterDB.Raw(`
		SELECT j.* FROM jobs j
		JOIN tenants t ON j.tenant_id = t.id
		WHERE j.status IN (?, ?) 
		AND (j.run_at IS NULL OR j.run_at <= NOW())
		AND t.status = 'active'
		ORDER BY j.priority DESC, j.created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`, models.JobStatusPending, models.JobStatusQueued).Scan(&job).Error
	
	if err != nil {
		return nil, err
	}
	
	return &job, nil
}

func (r *jobRepository) ClaimJob(jobID, workerID string) error {
	masterDB := r.masterDB.GetMasterDB()
	now := time.Now()
	
	result := masterDB.Exec(`
		UPDATE jobs 
		SET status = ?, started_at = ?, updated_at = ?
		WHERE id = ? AND status IN (?, ?)
	`, models.JobStatusRunning, now, now, jobID, models.JobStatusPending, models.JobStatusQueued)
	
	return result.Error
}

func (r *jobRepository) ReleaseJob(jobID string) error {
	masterDB := r.masterDB.GetMasterDB()
	now := time.Now()
	
	result := masterDB.Exec(`
		UPDATE jobs 
		SET status = ?, started_at = NULL, updated_at = ?
		WHERE id = ? AND status = ?
	`, models.JobStatusQueued, now, jobID, models.JobStatusRunning)
	
	return result.Error
}

func (r *jobRepository) GetJobsByStatus(tenantID string, status models.JobStatus) ([]*models.Job, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	rows, err := db.Raw(`
		SELECT id, tenant_id, type, status, priority, payload, result, error_message, retry_count, max_retries, run_at, started_at, completed_at, last_error, timeout_at, created_by, tags, dependencies, created_at, updated_at
		FROM jobs 
		WHERE tenant_id = ? AND status = ?
		ORDER BY priority DESC, created_at ASC
	`, tenantID, status).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var job models.Job
		var payloadJSON, resultJSON, tagsJSON, dependenciesJSON string
		
		rows.Scan(
			&job.ID, &job.TenantID, &job.Type, &job.Status, &job.Priority, &payloadJSON, &resultJSON,
			&job.ErrorMessage, &job.RetryCount, &job.MaxRetries, &job.RunAt, &job.StartedAt, &job.CompletedAt,
			&job.LastError, &job.TimeoutAt, &job.CreatedBy, &tagsJSON, &dependenciesJSON,
			&job.CreatedAt, &job.UpdatedAt,
		)

		json.Unmarshal([]byte(payloadJSON), &job.Payload)
		json.Unmarshal([]byte(resultJSON), &job.Result)
		json.Unmarshal([]byte(tagsJSON), &job.Tags)
		json.Unmarshal([]byte(dependenciesJSON), &job.Dependencies)

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func (r *jobRepository) GetReadyJobs(limit int) ([]*models.Job, error) {
	masterDB := r.masterDB.GetMasterDB()
	
	rows, err := masterDB.Raw(`
		SELECT j.* FROM jobs j
		JOIN tenants t ON j.tenant_id = t.id
		WHERE j.status IN (?, ?) 
		AND (j.run_at IS NULL OR j.run_at <= NOW())
		AND t.status = 'active'
		ORDER BY j.priority DESC, j.created_at ASC
		LIMIT ?
	`, models.JobStatusPending, models.JobStatusQueued, limit).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var job models.Job
		var payloadJSON, resultJSON, tagsJSON, dependenciesJSON string
		
		rows.Scan(
			&job.ID, &job.TenantID, &job.Type, &job.Status, &job.Priority, &payloadJSON, &resultJSON,
			&job.ErrorMessage, &job.RetryCount, &job.MaxRetries, &job.RunAt, &job.StartedAt, &job.CompletedAt,
			&job.LastError, &job.TimeoutAt, &job.CreatedBy, &tagsJSON, &dependenciesJSON,
			&job.CreatedAt, &job.UpdatedAt,
		)

		json.Unmarshal([]byte(payloadJSON), &job.Payload)
		json.Unmarshal([]byte(resultJSON), &job.Result)
		json.Unmarshal([]byte(tagsJSON), &job.Tags)
		json.Unmarshal([]byte(dependenciesJSON), &job.Dependencies)

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// Scheduled Job Methods
func (r *jobRepository) CreateScheduledJob(tenantID string, scheduledJob *models.ScheduledJob) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	payloadJSON, _ := json.Marshal(scheduledJob.Payload)

	result := db.Exec(`
		INSERT INTO scheduled_jobs (id, tenant_id, name, job_type, cron_expression, payload, priority, max_retries, is_active, next_run_at, last_run_at, last_job_id, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, scheduledJob.ID, scheduledJob.TenantID, scheduledJob.Name, scheduledJob.JobType, scheduledJob.CronExpression,
		string(payloadJSON), scheduledJob.Priority, scheduledJob.MaxRetries, scheduledJob.IsActive,
		scheduledJob.NextRunAt, scheduledJob.LastRunAt, scheduledJob.LastJobID, scheduledJob.CreatedBy,
		scheduledJob.CreatedAt, scheduledJob.UpdatedAt)

	return result.Error
}

func (r *jobRepository) GetScheduledJob(tenantID, scheduledJobID string) (*models.ScheduledJob, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	var scheduledJob models.ScheduledJob
	var payloadJSON string
	
	err = db.Raw(`
		SELECT id, tenant_id, name, job_type, cron_expression, payload, priority, max_retries, is_active, next_run_at, last_run_at, last_job_id, created_by, created_at, updated_at
		FROM scheduled_jobs 
		WHERE id = ? AND tenant_id = ?
	`, scheduledJobID, tenantID).Row().Scan(
		&scheduledJob.ID, &scheduledJob.TenantID, &scheduledJob.Name, &scheduledJob.JobType,
		&scheduledJob.CronExpression, &payloadJSON, &scheduledJob.Priority, &scheduledJob.MaxRetries,
		&scheduledJob.IsActive, &scheduledJob.NextRunAt, &scheduledJob.LastRunAt, &scheduledJob.LastJobID,
		&scheduledJob.CreatedBy, &scheduledJob.CreatedAt, &scheduledJob.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(payloadJSON), &scheduledJob.Payload)

	return &scheduledJob, nil
}

func (r *jobRepository) UpdateScheduledJob(tenantID string, scheduledJob *models.ScheduledJob) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	payloadJSON, _ := json.Marshal(scheduledJob.Payload)

	result := db.Exec(`
		UPDATE scheduled_jobs 
		SET name = ?, job_type = ?, cron_expression = ?, payload = ?, priority = ?, max_retries = ?, is_active = ?, next_run_at = ?, last_run_at = ?, last_job_id = ?, updated_at = ?
		WHERE id = ? AND tenant_id = ?
	`, scheduledJob.Name, scheduledJob.JobType, scheduledJob.CronExpression, string(payloadJSON),
		scheduledJob.Priority, scheduledJob.MaxRetries, scheduledJob.IsActive, scheduledJob.NextRunAt,
		scheduledJob.LastRunAt, scheduledJob.LastJobID, scheduledJob.UpdatedAt, scheduledJob.ID, tenantID)

	return result.Error
}

func (r *jobRepository) DeleteScheduledJob(tenantID, scheduledJobID string) error {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return err
	}

	result := db.Exec(`
		DELETE FROM scheduled_jobs 
		WHERE id = ? AND tenant_id = ?
	`, scheduledJobID, tenantID)

	return result.Error
}

func (r *jobRepository) ListScheduledJobs(tenantID string, limit, offset int) ([]*models.ScheduledJob, int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	var count int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM scheduled_jobs 
		WHERE tenant_id = ?
	`, tenantID).Scan(&count)

	// Get scheduled jobs
	rows, err := db.Raw(`
		SELECT id, tenant_id, name, job_type, cron_expression, payload, priority, max_retries, is_active, next_run_at, last_run_at, last_job_id, created_by, created_at, updated_at
		FROM scheduled_jobs 
		WHERE tenant_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, limit, offset).Rows()

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var scheduledJobs []*models.ScheduledJob
	for rows.Next() {
		var scheduledJob models.ScheduledJob
		var payloadJSON string
		
		rows.Scan(
			&scheduledJob.ID, &scheduledJob.TenantID, &scheduledJob.Name, &scheduledJob.JobType,
			&scheduledJob.CronExpression, &payloadJSON, &scheduledJob.Priority, &scheduledJob.MaxRetries,
			&scheduledJob.IsActive, &scheduledJob.NextRunAt, &scheduledJob.LastRunAt, &scheduledJob.LastJobID,
			&scheduledJob.CreatedBy, &scheduledJob.CreatedAt, &scheduledJob.UpdatedAt,
		)

		json.Unmarshal([]byte(payloadJSON), &scheduledJob.Payload)
		scheduledJobs = append(scheduledJobs, &scheduledJob)
	}

	return scheduledJobs, count, nil
}

func (r *jobRepository) GetDueScheduledJobs() ([]*models.ScheduledJob, error) {
	masterDB := r.masterDB.GetMasterDB()
	
	rows, err := masterDB.Raw(`
		SELECT sj.* FROM scheduled_jobs sj
		JOIN tenants t ON sj.tenant_id = t.id
		WHERE sj.is_active = true 
		AND sj.next_run_at <= NOW()
		AND t.status = 'active'
		ORDER BY sj.next_run_at ASC
	`).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scheduledJobs []*models.ScheduledJob
	for rows.Next() {
		var scheduledJob models.ScheduledJob
		var payloadJSON string
		
		rows.Scan(
			&scheduledJob.ID, &scheduledJob.TenantID, &scheduledJob.Name, &scheduledJob.JobType,
			&scheduledJob.CronExpression, &payloadJSON, &scheduledJob.Priority, &scheduledJob.MaxRetries,
			&scheduledJob.IsActive, &scheduledJob.NextRunAt, &scheduledJob.LastRunAt, &scheduledJob.LastJobID,
			&scheduledJob.CreatedBy, &scheduledJob.CreatedAt, &scheduledJob.UpdatedAt,
		)

		json.Unmarshal([]byte(payloadJSON), &scheduledJob.Payload)
		scheduledJobs = append(scheduledJobs, &scheduledJob)
	}

	return scheduledJobs, nil
}

func (r *jobRepository) GetJobMetrics(tenantID string, jobType string, dateFrom, dateTo *time.Time) (*models.JobMetrics, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	whereClause := "WHERE tenant_id = ?"
	args := []interface{}{tenantID}

	if jobType != "" {
		whereClause += " AND type = ?"
		args = append(args, jobType)
	}
	if dateFrom != nil {
		whereClause += " AND created_at >= ?"
		args = append(args, dateFrom)
	}
	if dateTo != nil {
		whereClause += " AND created_at <= ?"
		args = append(args, dateTo)
	}

	metrics := &models.JobMetrics{
		TenantID: tenantID,
		JobType:  jobType,
	}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_jobs,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_jobs,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_jobs,
			COUNT(CASE WHEN status = 'running' THEN 1 END) as running_jobs,
			COUNT(CASE WHEN status IN ('pending', 'queued') THEN 1 END) as queued_jobs,
			AVG(CASE WHEN completed_at IS NOT NULL AND started_at IS NOT NULL 
				THEN EXTRACT(EPOCH FROM (completed_at - started_at)) END) as avg_execution_time,
			SUM(CASE WHEN completed_at IS NOT NULL AND started_at IS NOT NULL 
				THEN EXTRACT(EPOCH FROM (completed_at - started_at)) END) as total_execution_time,
			MAX(completed_at) as last_executed
		FROM jobs %s
	`, whereClause)

	err = db.Raw(query, args...).Row().Scan(
		&metrics.TotalJobs, &metrics.CompletedJobs, &metrics.FailedJobs,
		&metrics.RunningJobs, &metrics.QueuedJobs, &metrics.AvgExecutionTime,
		&metrics.TotalExecutionTime, &metrics.LastExecuted,
	)

	if err != nil {
		return nil, err
	}

	if metrics.TotalJobs > 0 {
		metrics.SuccessRate = float64(metrics.CompletedJobs) / float64(metrics.TotalJobs) * 100
	}

	return metrics, nil
}

func (r *jobRepository) GetTenantJobStats(tenantID string) (map[string]int64, error) {
	db, err := r.tenantDBManager.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)

	rows, err := db.Raw(`
		SELECT status, COUNT(*) as count
		FROM jobs 
		WHERE tenant_id = ?
		GROUP BY status
	`, tenantID).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		rows.Scan(&status, &count)
		stats[status] = count
	}

	return stats, nil
}

func (r *jobRepository) CleanupOldJobs(olderThan time.Time) error {
	masterDB := r.masterDB.GetMasterDB()
	
	result := masterDB.Exec(`
		DELETE FROM jobs 
		WHERE status IN (?, ?) 
		AND completed_at < ?
	`, models.JobStatusCompleted, models.JobStatusFailed, olderThan)
	
	return result.Error
}