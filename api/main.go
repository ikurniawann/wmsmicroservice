package handler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
)

// In-memory storage for serverless
var (
	users = make(map[string]User)
	mu    sync.RWMutex
)

// Models
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	IsActive     bool      `json:"is_active"`
	IsVerified   bool      `json:"is_verified"`
	LastLoginAt  time.Time `json:"last_login_at"`
	CreatedAt    time.Time `json:"created_at"`
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// Initialize with default admin user
func init() {
	mu.Lock()
	defer mu.Unlock()
	
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	adminID := uuid.New().String()
	users["admin"] = User{
		ID:           adminID,
		Email:        "admin@wms.local",
		Username:     "admin",
		PasswordHash: string(hash),
		FirstName:    "System",
		LastName:     "Administrator",
		IsActive:     true,
		IsVerified:   true,
		CreatedAt:    time.Now(),
	}
	users[adminID] = users["admin"]
}

// JWT
var jwtSecret = []byte("wms-super-secret-key-for-jwt-signing-2026")

type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func generateToken(userID, username, email string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    []string{},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Handlers
func register(c echo.Context) error {
	type Request struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "All fields are required"})
	}

	mu.Lock()
	defer mu.Unlock()

	// Check if user exists
	if _, exists := users[req.Username]; exists {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Username already exists"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	user := User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
		IsActive:     true,
		IsVerified:   false,
		CreatedAt:    time.Now(),
	}

	users[user.Username] = user
	users[user.ID] = user

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]string{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
		},
	})
}

func login(c echo.Context) error {
	type Request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username and password are required"})
	}

	mu.RLock()
	user, exists := users[req.Username]
	mu.RUnlock()

	if !exists {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	if !user.CheckPassword(req.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	if !user.IsActive {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Account deactivated"})
	}

	mu.Lock()
	user.LastLoginAt = time.Now()
	users[req.Username] = user
	mu.Unlock()

	token, err := generateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   86400,
		"user": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
		},
	})
}

// Health check
func health(c echo.Context) error {
	mu.RLock()
	count := len(users)
	mu.RUnlock()
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":      "healthy",
		"service":     "auth",
		"db_status":   "in-memory",
		"user_count":  count,
	})
}

// Main Handler
func Handler(w http.ResponseWriter, r *http.Request) {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))
	e.Use(middleware.Recover())

	// Routes
	e.GET("/health", health)
	e.POST("/api/v1/auth/register", register)
	e.POST("/api/v1/auth/login", login)

	e.ServeHTTP(w, r)
}
