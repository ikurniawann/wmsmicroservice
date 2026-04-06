package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a system user
type User struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Email         string         `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	Username      string         `json:"username" gorm:"uniqueIndex;not null" validate:"required,min=3,max=50"`
	PasswordHash  string         `json:"-" gorm:"not null"`
	FirstName     string         `json:"first_name"`
	LastName      string         `json:"last_name"`
	Phone         string         `json:"phone"`
	IsActive      bool           `json:"is_active" gorm:"default:true"`
	IsVerified    bool           `json:"is_verified" gorm:"default:false"`
	LastLoginAt   *time.Time     `json:"last_login_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	Roles         []Role         `json:"roles,omitempty" gorm:"many2many:user_roles;"`
}

// SetPassword hashes and sets the user password
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies the password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// FullName returns the user's full name
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}
	return u.FirstName + " " + u.LastName
}

// Role represents a user role
type Role struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`
	Permissions JSONB     `json:"permissions" gorm:"type:jsonb;default:'[]'"`
	CreatedAt   time.Time `json:"created_at"`
	Users       []User    `json:"users,omitempty" gorm:"many2many:user_roles;"`
}

// JSONB type for PostgreSQL jsonb
type JSONB []byte

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"not null"`
	TokenHash string         `json:"-" gorm:"not null"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	RevokedAt *time.Time     `json:"revoked_at,omitempty"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// UserRole represents the junction table for many-to-many
type UserRole struct {
	UserID     uuid.UUID `json:"user_id" gorm:"primary_key"`
	RoleID     uuid.UUID `json:"role_id" gorm:"primary_key"`
	AssignedAt time.Time `json:"assigned_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// SeedRoles creates default roles in database
func SeedRoles(db *gorm.DB) error {
	roles := []Role{
		{
			Name:        "superadmin",
			Description: "Full system access",
			Permissions: JSONB(`["*"]`),
		},
		{
			Name:        "admin",
			Description: "Administrative access",
			Permissions: JSONB(`["users:read", "users:write", "products:read", "products:write", "inventory:read", "inventory:write"]`),
		},
		{
			Name:        "warehouse_manager",
			Description: "Warehouse management",
			Permissions: JSONB(`["inventory:read", "inventory:write", "products:read", "reports:read"]`),
		},
		{
			Name:        "warehouse_staff",
			Description: "Warehouse operations",
			Permissions: JSONB(`["inventory:read", "inventory:write:limited"]`),
		},
		{
			Name:        "cashier",
			Description: "POS operations",
			Permissions: JSONB(`["pos:read", "pos:write", "products:read"]`),
		},
		{
			Name:        "viewer",
			Description: "Read-only access",
			Permissions: JSONB(`["products:read", "inventory:read", "reports:read"]`),
		},
	}

	for _, role := range roles {
		var existing Role
		if err := db.Where("name = ?", role.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&role).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}

// SeedAdmin creates default admin user
func SeedAdmin(db *gorm.DB) error {
	var existing User
	if err := db.Where("username = ?", "admin").First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			admin := User{
				Email:      "admin@wms.local",
				Username:   "admin",
				FirstName:  "System",
				LastName:   "Administrator",
				IsActive:   true,
				IsVerified: true,
			}
			
			if err := admin.SetPassword("admin123"); err != nil {
				return err
			}
			
			if err := db.Create(&admin).Error; err != nil {
				return err
			}
			
			// Assign superadmin role
			var superadminRole Role
			if err := db.Where("name = ?", "superadmin").First(&superadminRole).Error; err != nil {
				return err
			}
			
			if err := db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", admin.ID, superadminRole.ID).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
