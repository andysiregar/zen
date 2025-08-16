package services

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/zen/shared/pkg/models"
	"ticket-service/internal/repositories"
)

type TicketService interface {
	// Ticket CRUD
	CreateTicket(userID, tenantID string, req *models.TicketCreateRequest) (*models.TicketResponse, error)
	GetTicket(userID, tenantID, ticketID string) (*models.TicketResponse, error)
	UpdateTicket(userID, tenantID, ticketID string, req *models.TicketUpdateRequest) (*models.TicketResponse, error)
	DeleteTicket(userID, tenantID, ticketID string) error
	ListTickets(userID, tenantID string, limit, offset int, filters repositories.TicketFilters) ([]*models.TicketResponse, int64, error)
	
	// Ticket operations
	AssignTicket(userID, tenantID, ticketID, assigneeID string) (*models.TicketResponse, error)
	UpdateTicketStatus(userID, tenantID, ticketID string, status models.TicketStatus) (*models.TicketResponse, error)
	UpdateTicketPriority(userID, tenantID, ticketID string, priority models.TicketPriority) (*models.TicketResponse, error)
	
	// Comments
	CreateComment(userID, tenantID, ticketID string, req *models.TicketCommentCreateRequest) (*models.TicketCommentResponse, error)
	GetComments(userID, tenantID, ticketID string, includeInternal bool) ([]*models.TicketCommentResponse, error)
	UpdateComment(userID, tenantID, commentID string, content string) (*models.TicketCommentResponse, error)
	DeleteComment(userID, tenantID, commentID string) error
	
	// Search and stats
	SearchTickets(userID, tenantID, query string, limit, offset int) ([]*models.TicketResponse, int64, error)
	GetTicketStats(userID, tenantID string, dateFrom, dateTo *time.Time) (repositories.TicketStats, error)
}

type ticketService struct {
	repo   repositories.TicketRepository
	logger *zap.Logger
}

func NewTicketService(repo repositories.TicketRepository, logger *zap.Logger) TicketService {
	return &ticketService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ticketService) CreateTicket(userID, tenantID string, req *models.TicketCreateRequest) (*models.TicketResponse, error) {
	// Business logic validation
	if req.Title == "" {
		return nil, fmt.Errorf("ticket title is required")
	}
	if req.Description == "" {
		return nil, fmt.Errorf("ticket description is required")
	}

	// Set default values
	ticket := &models.Ticket{
		TenantID:    tenantID,
		Title:       req.Title,
		Description: req.Description,
		Status:      models.TicketStatusOpen,
		Priority:    models.TicketPriorityMedium,
		Type:        models.TicketTypeSupport,
		ReporterID:  userID,
		Category:    req.Category,
		ProjectID:   req.ProjectID,
	}

	// Override defaults if provided
	if req.Priority != "" {
		ticket.Priority = req.Priority
	}
	if req.Type != "" {
		ticket.Type = req.Type
	}

	// Set due date based on priority if not provided
	if req.DueDate != nil {
		ticket.DueDate = req.DueDate
	} else {
		ticket.DueDate = s.calculateDueDate(ticket.Priority)
	}

	// Convert tags to JSONB
	if req.Tags != nil && len(req.Tags) > 0 {
		// In production, properly serialize tags to JSONB
		ticket.Tags = make(models.JSONB)
	}

	// Create ticket in database
	err := s.repo.Create(tenantID, ticket)
	if err != nil {
		s.logger.Error("Failed to create ticket", zap.Error(err), zap.String("tenant_id", tenantID))
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	s.logger.Info("Ticket created successfully", 
		zap.String("ticket_id", ticket.ID),
		zap.String("tenant_id", tenantID),
		zap.String("reporter_id", userID))

	// TODO: Send notifications (email, webhooks, etc.)
	// TODO: Create activity log entry

	response := ticket.ToResponse()
	return &response, nil
}

func (s *ticketService) GetTicket(userID, tenantID, ticketID string) (*models.TicketResponse, error) {
	ticket, err := s.repo.GetByID(tenantID, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	// Business logic: Check if user can view this ticket
	if !s.canUserViewTicket(userID, ticket) {
		return nil, fmt.Errorf("access denied: user cannot view this ticket")
	}

	response := ticket.ToResponse()
	return &response, nil
}

func (s *ticketService) UpdateTicket(userID, tenantID, ticketID string, req *models.TicketUpdateRequest) (*models.TicketResponse, error) {
	// Get existing ticket
	ticket, err := s.repo.GetByID(tenantID, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	// Business logic: Check if user can update this ticket
	if !s.canUserUpdateTicket(userID, ticket) {
		return nil, fmt.Errorf("access denied: user cannot update this ticket")
	}

	// Update fields if provided
	originalStatus := ticket.Status
	if req.Title != nil {
		ticket.Title = *req.Title
	}
	if req.Description != nil {
		ticket.Description = *req.Description
	}
	if req.Status != nil {
		ticket.Status = *req.Status
		// Set resolved/closed timestamps
		if *req.Status == models.TicketStatusResolved && ticket.ResolvedAt == nil {
			now := time.Now()
			ticket.ResolvedAt = &now
		}
		if *req.Status == models.TicketStatusClosed && ticket.ClosedAt == nil {
			now := time.Now()
			ticket.ClosedAt = &now
		}
	}
	if req.Priority != nil {
		ticket.Priority = *req.Priority
		// Recalculate due date if priority changed
		if ticket.DueDate != nil {
			ticket.DueDate = s.calculateDueDate(ticket.Priority)
		}
	}
	if req.Type != nil {
		ticket.Type = *req.Type
	}
	if req.Category != nil {
		ticket.Category = *req.Category
	}
	if req.AssigneeID != nil {
		ticket.AssigneeID = *req.AssigneeID
	}
	if req.ProjectID != nil {
		ticket.ProjectID = *req.ProjectID
	}
	if req.DueDate != nil {
		ticket.DueDate = req.DueDate
	}

	// Save updated ticket
	err = s.repo.Update(tenantID, ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}

	s.logger.Info("Ticket updated successfully", 
		zap.String("ticket_id", ticketID),
		zap.String("tenant_id", tenantID),
		zap.String("updated_by", userID))

	// TODO: Send notifications if status changed
	if req.Status != nil && *req.Status != originalStatus {
		s.logger.Info("Ticket status changed", 
			zap.String("ticket_id", ticketID),
			zap.String("old_status", string(originalStatus)),
			zap.String("new_status", string(*req.Status)))
		// TODO: Send status change notifications
	}

	response := ticket.ToResponse()
	return &response, nil
}

func (s *ticketService) DeleteTicket(userID, tenantID, ticketID string) error {
	// Get existing ticket to check permissions
	ticket, err := s.repo.GetByID(tenantID, ticketID)
	if err != nil {
		return fmt.Errorf("failed to get ticket: %w", err)
	}

	// Business logic: Check if user can delete this ticket
	if !s.canUserDeleteTicket(userID, ticket) {
		return fmt.Errorf("access denied: user cannot delete this ticket")
	}

	err = s.repo.Delete(tenantID, ticketID)
	if err != nil {
		return fmt.Errorf("failed to delete ticket: %w", err)
	}

	s.logger.Info("Ticket deleted successfully", 
		zap.String("ticket_id", ticketID),
		zap.String("tenant_id", tenantID),
		zap.String("deleted_by", userID))

	return nil
}

func (s *ticketService) ListTickets(userID, tenantID string, limit, offset int, filters repositories.TicketFilters) ([]*models.TicketResponse, int64, error) {
	// Business logic: Apply user-specific filters
	// For example, regular users might only see their own tickets
	// TODO: Check user role and apply appropriate filters

	tickets, total, err := s.repo.List(tenantID, limit, offset, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tickets: %w", err)
	}

	// Convert to response format
	var responses []*models.TicketResponse
	for _, ticket := range tickets {
		if s.canUserViewTicket(userID, ticket) {
			response := ticket.ToResponse()
			responses = append(responses, &response)
		}
	}

	return responses, total, nil
}

func (s *ticketService) AssignTicket(userID, tenantID, ticketID, assigneeID string) (*models.TicketResponse, error) {
	ticket, err := s.repo.GetByID(tenantID, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	// Business logic: Check if user can assign tickets
	if !s.canUserAssignTickets(userID) {
		return nil, fmt.Errorf("access denied: user cannot assign tickets")
	}

	// TODO: Validate that assigneeID is a valid user in this tenant

	ticket.AssigneeID = assigneeID
	err = s.repo.Update(tenantID, ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to assign ticket: %w", err)
	}

	s.logger.Info("Ticket assigned successfully", 
		zap.String("ticket_id", ticketID),
		zap.String("assignee_id", assigneeID),
		zap.String("assigned_by", userID))

	// TODO: Send assignment notification to assignee

	response := ticket.ToResponse()
	return &response, nil
}

func (s *ticketService) UpdateTicketStatus(userID, tenantID, ticketID string, status models.TicketStatus) (*models.TicketResponse, error) {
	return s.UpdateTicket(userID, tenantID, ticketID, &models.TicketUpdateRequest{
		Status: &status,
	})
}

func (s *ticketService) UpdateTicketPriority(userID, tenantID, ticketID string, priority models.TicketPriority) (*models.TicketResponse, error) {
	return s.UpdateTicket(userID, tenantID, ticketID, &models.TicketUpdateRequest{
		Priority: &priority,
	})
}

// Comment methods
func (s *ticketService) CreateComment(userID, tenantID, ticketID string, req *models.TicketCommentCreateRequest) (*models.TicketCommentResponse, error) {
	// Verify ticket exists and user can comment
	ticket, err := s.repo.GetByID(tenantID, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	if !s.canUserViewTicket(userID, ticket) {
		return nil, fmt.Errorf("access denied: user cannot comment on this ticket")
	}

	comment := &models.TicketComment{
		TicketID:   ticketID,
		AuthorID:   userID,
		Content:    req.Content,
		IsInternal: req.IsInternal,
	}

	err = s.repo.CreateComment(tenantID, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	s.logger.Info("Comment created successfully", 
		zap.String("ticket_id", ticketID),
		zap.String("author_id", userID),
		zap.Bool("is_internal", req.IsInternal))

	response := comment.ToResponse()
	return &response, nil
}

func (s *ticketService) GetComments(userID, tenantID, ticketID string, includeInternal bool) ([]*models.TicketCommentResponse, error) {
	// Verify ticket exists and user can view
	ticket, err := s.repo.GetByID(tenantID, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	if !s.canUserViewTicket(userID, ticket) {
		return nil, fmt.Errorf("access denied: user cannot view comments on this ticket")
	}

	// Business logic: Only agents/admins can see internal comments
	if includeInternal && !s.canUserViewInternalComments(userID) {
		includeInternal = false
	}

	comments, err := s.repo.GetComments(tenantID, ticketID, includeInternal)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	var responses []*models.TicketCommentResponse
	for _, comment := range comments {
		response := comment.ToResponse()
		responses = append(responses, &response)
	}

	return responses, nil
}

func (s *ticketService) UpdateComment(userID, tenantID, commentID string, content string) (*models.TicketCommentResponse, error) {
	// TODO: Implement comment update logic
	return nil, fmt.Errorf("comment update not implemented yet")
}

func (s *ticketService) DeleteComment(userID, tenantID, commentID string) error {
	// TODO: Implement comment deletion logic
	return fmt.Errorf("comment deletion not implemented yet")
}

func (s *ticketService) SearchTickets(userID, tenantID, query string, limit, offset int) ([]*models.TicketResponse, int64, error) {
	tickets, total, err := s.repo.Search(tenantID, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search tickets: %w", err)
	}

	var responses []*models.TicketResponse
	for _, ticket := range tickets {
		if s.canUserViewTicket(userID, ticket) {
			response := ticket.ToResponse()
			responses = append(responses, &response)
		}
	}

	return responses, total, nil
}

func (s *ticketService) GetTicketStats(userID, tenantID string, dateFrom, dateTo *time.Time) (repositories.TicketStats, error) {
	// TODO: Check if user can view stats (admin/manager role)
	return s.repo.GetTicketStats(tenantID, dateFrom, dateTo)
}

// Business logic helper methods
func (s *ticketService) calculateDueDate(priority models.TicketPriority) *time.Time {
	now := time.Now()
	var dueDate time.Time

	switch priority {
	case models.TicketPriorityCritical:
		dueDate = now.Add(4 * time.Hour) // 4 hours for critical
	case models.TicketPriorityHigh:
		dueDate = now.Add(24 * time.Hour) // 1 day for high
	case models.TicketPriorityMedium:
		dueDate = now.Add(72 * time.Hour) // 3 days for medium
	case models.TicketPriorityLow:
		dueDate = now.Add(168 * time.Hour) // 7 days for low
	default:
		dueDate = now.Add(72 * time.Hour) // Default to 3 days
	}

	return &dueDate
}

func (s *ticketService) canUserViewTicket(userID string, ticket *models.Ticket) bool {
	// Basic permission: user can view tickets they created or are assigned to
	// TODO: Implement proper role-based permissions
	return ticket.ReporterID == userID || ticket.AssigneeID == userID
}

func (s *ticketService) canUserUpdateTicket(userID string, ticket *models.Ticket) bool {
	// Basic permission: user can update tickets they created or are assigned to
	// TODO: Implement proper role-based permissions
	return ticket.ReporterID == userID || ticket.AssigneeID == userID
}

func (s *ticketService) canUserDeleteTicket(userID string, ticket *models.Ticket) bool {
	// Restrictive permission: only ticket creator can delete (or admin)
	// TODO: Implement proper role-based permissions
	return ticket.ReporterID == userID
}

func (s *ticketService) canUserAssignTickets(userID string) bool {
	// TODO: Check if user has agent/admin role
	// For now, allow all users to assign tickets
	return true
}

func (s *ticketService) canUserViewInternalComments(userID string) bool {
	// TODO: Check if user has agent/admin role
	// For now, allow all users to view internal comments
	return true
}