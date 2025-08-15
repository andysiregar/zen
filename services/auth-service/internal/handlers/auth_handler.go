package handlers

import (
	"net/url"
	"github.com/gin-gonic/gin"
	"github.com/zen/shared/pkg/auth"
	"github.com/zen/shared/pkg/models"
	"github.com/zen/shared/pkg/utils"
	"github.com/zen/shared/pkg/routing"
	"auth-service/internal/services"
)

type AuthHandler struct {
	authService     services.AuthService
	jwtService      *auth.JWTService
	regionalRouter  *routing.RegionalRouter
}

func NewAuthHandler(authService services.AuthService, jwtService *auth.JWTService, regionalRouter *routing.RegionalRouter) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		jwtService:     jwtService,
		regionalRouter: regionalRouter,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	OrganizationID string `json:"organization_id" binding:"required,uuid"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=8"`
	FirstName      string `json:"first_name" binding:"required,min=2,max=100"`
	LastName       string `json:"last_name" binding:"required,min=2,max=100"`
	Phone          string `json:"phone,omitempty"`
	Timezone       string `json:"timezone,omitempty"`
	Language       string `json:"language,omitempty"`
}

type AuthResponse struct {
	User         *models.UserResponse `json:"user"`
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	ExpiresIn    int                  `json:"expires_in"` // seconds
	RedirectURL  string               `json:"redirect_url,omitempty"` // Regional redirect after auth
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// Authenticate user
	user, err := h.authService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid email or password")
		return
	}

	// Generate JWT tokens
	tokenPair, err := h.jwtService.GenerateTokenPair(
		user.ID,
		"", // TenantID will be set when user selects a tenant
		user.OrganizationID,
		user.Email,
		string(user.Role),
		[]string{}, // TODO: Implement permissions
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate tokens")
		return
	}

	response := AuthResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    24 * 3600, // 24 hours in seconds
	}

	utils.SuccessResponse(c, response, "Login successful")
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// Create user
	createUserReq := models.UserCreateRequest{
		OrganizationID: req.OrganizationID,
		Email:          req.Email,
		Password:       req.Password,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Phone:          req.Phone,
		Timezone:       req.Timezone,
		Language:       req.Language,
	}

	user, err := h.authService.CreateUser(createUserReq)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create user: "+err.Error())
		return
	}

	// Generate JWT tokens
	tokenPair, err := h.jwtService.GenerateTokenPair(
		user.ID,
		"", // TenantID will be set when user selects a tenant
		user.OrganizationID,
		user.Email,
		string(user.Role),
		[]string{}, // TODO: Implement permissions
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate tokens")
		return
	}

	response := AuthResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    24 * 3600, // 24 hours in seconds
	}

	utils.CreatedResponse(c, response, "User registered successfully")
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// Refresh access token
	newAccessToken, err := h.jwtService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid or expired refresh token")
		return
	}

	response := gin.H{
		"access_token": newAccessToken,
		"expires_in":   24 * 3600, // 24 hours in seconds
	}

	utils.SuccessResponse(c, response, "Token refreshed successfully")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT system, logout is typically handled client-side
	// by simply discarding the tokens. However, you could implement
	// token blacklisting here if needed.
	
	utils.SuccessResponse(c, gin.H{"message": "Logged out successfully"}, "Success")
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from JWT context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	user, err := h.authService.GetUserByID(userID.(string))
	if err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	utils.SuccessResponse(c, user, "Profile retrieved successfully")
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	user, err := h.authService.UpdateUser(userID.(string), req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update profile: "+err.Error())
		return
	}

	utils.SuccessResponse(c, user, "Profile updated successfully")
}

// LoginWithRedirect handles authentication with regional routing support
// Used when users come from custom domains and need to be redirected to their regional portal
func (h *AuthHandler) LoginWithRedirect(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body: "+err.Error())
		return
	}

	// Get redirect parameters
	returnURL := c.Query("return_url")
	// customDomain := c.Query("custom_domain") // Reserved for future use

	// Authenticate user
	user, err := h.authService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid email or password")
		return
	}

	// If return URL is provided, extract tenant information and redirect
	var redirectURL string
	if returnURL != "" {
		// Decode the return URL
		decodedURL, err := url.QueryUnescape(returnURL)
		if err == nil {
			redirectURL = decodedURL
		}
	}

	// Generate JWT tokens
	tokenPair, err := h.jwtService.GenerateTokenPair(
		user.ID,
		"", // TenantID will be set when user accesses tenant portal
		user.OrganizationID,
		user.Email,
		string(user.Role),
		[]string{}, // TODO: Implement permissions
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate tokens")
		return
	}

	response := AuthResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    24 * 3600, // 24 hours in seconds
		RedirectURL:  redirectURL,
	}

	utils.SuccessResponse(c, response, "Login successful")
}

// HandleCustomDomain handles requests from custom domains
// Redirects to appropriate regional authentication
func (h *AuthHandler) HandleCustomDomain(c *gin.Context) {
	host := c.Request.Host
	
	// Check if it's a custom domain
	isCustom, domain := h.regionalRouter.ParseCustomDomain(host)
	if !isCustom {
		utils.BadRequestResponse(c, "Not a custom domain")
		return
	}

	// TODO: Look up tenant by custom domain to get region info
	// For now, we'll return instructions for the client to handle
	utils.SuccessResponse(c, gin.H{
		"custom_domain": domain,
		"auth_url": "https://chilldesk.io/auth/login",
		"message": "Please authenticate through the main portal",
	}, "Custom domain detected")
}