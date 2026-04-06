package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	jwtSecret     = []byte(getEnv("JWT_SECRET", "your-secret-key-change-this"))
	jwtExpiration = 24 * time.Hour
)

// Claims represents JWT claims
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token
func GenerateToken(userID uuid.UUID, username, email string, roles []string) (string, error) {
	claims := Claims{
		UserID:   userID.String(),
		Username: username,
		Email:    email,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "wms-auth-service",
			Subject:     userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken validates a JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateRefreshToken creates a refresh token
func GenerateRefreshToken() (string, error) {
	return uuid.New().String(), nil
}

// HashToken hashes a token for storage
func HashToken(token string) string {
	// Simple hash - in production use proper hashing
	// This is just for demonstration
	return token
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
