package server

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// accessClaims holds access token claims (user_id, roles, permissions).
type accessClaims struct {
	jwt.RegisteredClaims
	UserID      string   `json:"user_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// refreshClaims holds refresh token claims (user_id only).
type refreshClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

// jwtConfig holds JWT secret and expirations (from env).
type jwtConfig struct {
	secret         []byte
	accessExpiry   time.Duration
	refreshExpiry  time.Duration
}

func loadJWTConfig() jwtConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "change-me-in-production"
	}
	accessMin := getIntEnv("JWT_ACCESS_EXPIRY_MINUTES", 15)
	refreshHours := getIntEnv("JWT_REFRESH_EXPIRY_HOURS", 24*7) // 7 days
	return jwtConfig{
		secret:        []byte(secret),
		accessExpiry:  time.Duration(accessMin) * time.Minute,
		refreshExpiry: time.Duration(refreshHours) * time.Hour,
	}
}

func getIntEnv(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// GenerateAccessToken returns a signed JWT access token with user_id, roles, permissions.
func (s *AuthServer) GenerateAccessToken(userID string, roles, permissions []string) (string, error) {
	cfg := loadJWTConfig()
	now := time.Now()
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.accessExpiry)),
		},
		UserID:      userID,
		Roles:       roles,
		Permissions: permissions,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.secret)
}

// GenerateRefreshToken returns a signed JWT refresh token with user_id.
func (s *AuthServer) GenerateRefreshToken(userID string) (string, error) {
	cfg := loadJWTConfig()
	now := time.Now()
	claims := refreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.refreshExpiry)),
		},
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.secret)
}

// ParseAccessToken verifies the access token and returns user_id, roles, permissions.
func (s *AuthServer) ParseAccessToken(tokenString string) (userID string, roles, permissions []string, err error) {
	cfg := loadJWTConfig()
	token, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(*jwt.Token) (interface{}, error) {
		return cfg.secret, nil
	})
	if err != nil {
		return "", nil, nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return "", nil, nil, fmt.Errorf("invalid token claims")
	}
	return claims.UserID, claims.Roles, claims.Permissions, nil
}
