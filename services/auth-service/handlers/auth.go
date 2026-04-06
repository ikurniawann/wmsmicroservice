package handlers

import (
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ikurniawann/wmsmicroservice/services/auth-service/config"
	"github.com/ikurniawann/wmsmicroservice/services/auth-service/models"
	"github.com/ikurniawann/wmsmicroservice/services/auth-service/utils"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var validate = validator.New()

// RegisterRequest represents registration request
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	User         UserInfo    `json:"user"`
}

// UserInfo represents user info in response
type UserInfo struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	Username  string   `json:"username"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Roles     []string `json:"roles"`
}

// Register handles user registration
func Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate
	if err := validate.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Check if user exists
	var existingUser models.User
	if err := config.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Database error",
			})
		}
	} else {
		return c.JSON(http.StatusConflict, map[string]string{
			"error": "User already exists",
		})
	}

	// Create user
	user := models.User{
		Email:     req.Email,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		IsActive:  true,
	}

	// Hash password
	if err := user.SetPassword(req.Password); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process password",
		})
	}

	// Save to database
	if err := config.DB.Create(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
	}

	// Assign default role (viewer)
	var viewerRole models.Role
	if err := config.DB.Where("name = ?", "viewer").First(&viewerRole).Error; err == nil {
		config.DB.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user.ID, viewerRole.ID)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user": UserInfo{
			ID:        user.ID.String(),
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Roles:     []string{"viewer"},
		},
	})
}

// Login handles user login
func Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate
	if err := validate.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Find user
	var user models.User
	if err := config.DB.Where("username = ? OR email = ?", req.Username, req.Username).Preload("Roles").First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid credentials",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database error",
		})
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid credentials",
		})
	}

	// Check if active
	if !user.IsActive {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Account is deactivated",
		})
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	config.DB.Save(&user)

	// Extract roles
	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}
	if len(roles) == 0 {
		roles = []string{"viewer"}
	}

	// Generate tokens
	accessToken, err := utils.GenerateToken(user.ID, user.Username, user.Email, roles)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate token",
		})
	}

	refreshToken, _ := utils.GenerateRefreshToken()

	// Save refresh token (simplified)
	// In production: hash and store in database

	return c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    86400, // 24 hours
		User: UserInfo{
			ID:        user.ID.String(),
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Roles:     roles,
		},
	})
}

// Logout handles user logout
func Logout(c echo.Context) error {
	// In production: invalidate refresh token
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// Refresh handles token refresh
func Refresh(c echo.Context) error {
	// Simplified - in production: validate refresh token
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Token refreshed",
	})
}

// Me returns current user info
func Me(c echo.Context) error {
	userID := c.Get("userID").(string)
	username := c.Get("username").(string)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":  userID,
		"username": username,
	})
}
