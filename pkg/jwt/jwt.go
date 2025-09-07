package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTConfig holds the JWT configuration
type JWTConfig struct {
	SecretKey      string
	AccessTokenExp time.Duration
	RefreshTokenExp time.Duration
}

// Claims represents the JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	SecretKey       []byte
	accessTokenExp  time.Duration
	refreshTokenExp time.Duration
}

// NewJWTManager creates a new JWTManager
func NewJWTManager(config *JWTConfig) *JWTManager {
	return &JWTManager{
		SecretKey:       []byte(config.SecretKey),
		accessTokenExp:  config.AccessTokenExp,
		refreshTokenExp: config.RefreshTokenExp,
	}
}

// GenerateAccessToken generates an access token for a user
func (j *JWTManager) GenerateAccessToken(userID, username, role string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SecretKey)
}

// GenerateRefreshToken generates a refresh token for a user
func (j *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshTokenExp)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Subject:   userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SecretKey)
}

// VerifyToken verifies a JWT token and returns the claims
func (j *JWTManager) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetUserIDFromToken extracts the user ID from a JWT token
func (j *JWTManager) GetUserIDFromToken(tokenString string) (string, error) {
	claims, err := j.VerifyToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// RefreshToken verifies a refresh token and generates new access and refresh tokens
func (j *JWTManager) RefreshToken(refreshTokenString string) (string, string, error) {
	token, err := jwt.ParseWithClaims(refreshTokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SecretKey, nil
	})

	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		// Generate new tokens
		// Note: In a real implementation, you would get the user details from the database
		accessToken, err := j.GenerateAccessToken(claims.Subject, "", "")
		if err != nil {
			return "", "", err
		}

		newRefreshToken, err := j.GenerateRefreshToken(claims.Subject)
		if err != nil {
			return "", "", err
		}

		return accessToken, newRefreshToken, nil
	}

	return "", "", errors.New("invalid refresh token")
}