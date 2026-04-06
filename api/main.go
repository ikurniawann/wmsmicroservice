package handler

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Models
type User struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Email        string     `json:"email" gorm:"uniqueIndex;not null"`
	Username     string     `json:"username" gorm:"uniqueIndex;not null"`
	PasswordHash string     `json:"-" gorm:"not null"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Phone        string     `json:"phone"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	IsVerified   bool       `json:"is_verified" gorm:"default:false"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

type Role struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Database
var db *gorm.DB

func initDB() error {
	// Supabase connection string
	dsn := "postgresql://postgres:empatTH3010*#@db.bnpvryotcgvposlbbcbd.supabase.co:5432/postgres?sslmode=require"

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database connection error: %v\n", err)
		return fmt.Errorf("database connection failed: %v", err)
	}
	
	fmt.Println("Database connected successfully")
	
	// Auto migrate
	if err := db.AutoMigrate(&User{}, &Role{}); err != nil {
		fmt.Fprintf(os.Stderr, "AutoMigrate error: %v\n", err)
		return fmt.Errorf("auto migrate failed: %v", err)
	}
	
	fmt.Println("AutoMigrate completed")
	
	return nil
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

func generateToken(userID uuid.UUID, username, email string) (string, error) {
	claims := Claims{
		UserID:   userID.String(),
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

	// Validate
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "All fields are required"})
	}

	if db == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database not initialized"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Password hash error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	user := User{
		ID:           uuid.New(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
		IsActive:     true,
		IsVerified:   false,
	}

	if err := db.Create(&user).Error; err != nil {
		fmt.Fprintf(os.Stderr, "Create user error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to create user: %v", err)})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]string{
			"id":       user.ID.String(),
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

	if db == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database not initialized"})
	}

	var user User
	if err := db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	if !user.CheckPassword(req.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	if !user.IsActive {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Account deactivated"})
	}

	now := time.Now()
	user.LastLoginAt = &now
	db.Save(&user)

	token, err := generateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   86400,
		"user": map[string]string{
			"id":       user.ID.String(),
			"email":    user.Email,
			"username": user.Username,
		},
	})
}

// Health check
func health(c echo.Context) error {
	dbStatus := "connected"
	if db == nil {
		dbStatus = "disconnected"
	}
	return c.JSON(http.StatusOK, map[string]string{
		"status":     "healthy",
		"service":    "auth",
		"db_status":  dbStatus,
	})
}

// Main Handler
func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize DB if not already done
	if db == nil {
		if err := initDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Database init error: %v\n", err)
			// Continue without DB - health check still works
		}
	}

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
