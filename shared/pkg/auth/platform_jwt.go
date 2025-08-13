package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zen/shared/pkg/models"
)

// PlatformClaims represents JWT claims for platform admin authentication
type PlatformClaims struct {
	AdminID     string                        `json:"admin_id"`
	Email       string                        `json:"email"`
	Role        models.PlatformRole           `json:"role"`
	Permissions []models.PlatformPermission   `json:"permissions"`
	jwt.RegisteredClaims
}

// PlatformJWTService handles JWT operations for platform admin authentication
type PlatformJWTService struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewPlatformJWTService creates a new platform JWT service
func NewPlatformJWTService(secretKey string, accessTokenTTL, refreshTokenTTL time.Duration) *PlatformJWTService {
	return &PlatformJWTService{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// GenerateTokens generates access and refresh tokens for platform admin
func (s *PlatformJWTService) GenerateTokens(admin *models.PlatformAdmin) (string, string, error) {
	// Access token
	accessClaims := &PlatformClaims{
		AdminID:     admin.ID,
		Email:       admin.Email,
		Role:        admin.Role,
		Permissions: admin.GetAllPermissions(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "zen-platform-auth",
			Subject:   admin.ID,
			ID:        fmt.Sprintf("access-%d", time.Now().Unix()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token
	refreshClaims := &PlatformClaims{
		AdminID: admin.ID,
		Email:   admin.Email,
		Role:    admin.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "zen-platform-auth",
			Subject:   admin.ID,
			ID:        fmt.Sprintf("refresh-%d", time.Now().Unix()),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessTokenString, refreshTokenString, nil
}

// ValidateToken validates and parses a platform admin JWT token
func (s *PlatformJWTService) ValidateToken(tokenString string) (*PlatformClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &PlatformClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*PlatformClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// RefreshToken generates a new access token from a valid refresh token
func (s *PlatformJWTService) RefreshToken(refreshTokenString string, admin *models.PlatformAdmin) (string, error) {
	claims, err := s.ValidateToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.AdminID != admin.ID {
		return "", fmt.Errorf("token admin ID mismatch")
	}

	// Generate new access token
	accessClaims := &PlatformClaims{
		AdminID:     admin.ID,
		Email:       admin.Email,
		Role:        admin.Role,
		Permissions: admin.GetAllPermissions(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "zen-platform-auth",
			Subject:   admin.ID,
			ID:        fmt.Sprintf("access-%d", time.Now().Unix()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	return accessToken.SignedString(s.secretKey)
}

// GetAccessTokenTTL returns the access token TTL in seconds
func (s *PlatformJWTService) GetAccessTokenTTL() int {
	return int(s.accessTokenTTL.Seconds())
}