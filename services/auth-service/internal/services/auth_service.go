package services

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"github.com/zen/shared/pkg/models"
	"auth-service/internal/repositories"
)

type AuthService interface {
	AuthenticateUser(email, password string) (*models.UserResponse, error)
	CreateUser(req models.UserCreateRequest) (*models.UserResponse, error)
	GetUserByID(userID string) (*models.UserResponse, error)
	UpdateUser(userID string, req models.UserUpdateRequest) (*models.UserResponse, error)
	GetUserByEmail(email string) (*models.UserResponse, error)
}

type authService struct {
	userRepo repositories.UserRepository
}

func NewAuthService(userRepo repositories.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

func (s *authService) AuthenticateUser(email, password string) (*models.UserResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if !user.IsActive() {
		return nil, fmt.Errorf("user account is not active")
	}

	// Check if account is locked
	if user.IsLocked() {
		return nil, fmt.Errorf("user account is locked")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Increment login attempts
		s.userRepo.IncrementLoginAttempts(user.ID)
		return nil, fmt.Errorf("invalid password")
	}

	// Reset login attempts on successful authentication
	s.userRepo.ResetLoginAttempts(user.ID)

	// Update last login
	s.userRepo.UpdateLastLogin(user.ID)

	response := user.ToResponse()
	return &response, nil
}

func (s *authService) CreateUser(req models.UserCreateRequest) (*models.UserResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		OrganizationID: req.OrganizationID,
		Email:          req.Email,
		Password:       string(hashedPassword),
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Phone:          req.Phone,
		Timezone:       req.Timezone,
		Language:       req.Language,
		Role:           models.RoleUser, // Default role
		Status:         models.StatusPending, // Pending email verification
	}

	// Set default values
	if user.Timezone == "" {
		user.Timezone = "UTC"
	}
	if user.Language == "" {
		user.Language = "en"
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *authService) GetUserByID(userID string) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *authService) UpdateUser(userID string, req models.UserUpdateRequest) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields if provided
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}
	if req.Language != nil {
		user.Language = *req.Language
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	err = s.userRepo.Update(user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *authService) GetUserByEmail(email string) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}