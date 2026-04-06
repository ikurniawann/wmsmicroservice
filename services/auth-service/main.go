package main

import (
	"log"
	"os"

	"github.com/ikurniawann/wmsmicroservice/services/auth-service/config"
	"github.com/ikurniawann/wmsmicroservice/services/auth-service/handlers"
	"github.com/ikurniawann/wmsmicroservice/services/auth-service/models"
	"github.com/ikurniawann/wmsmicroservice/services/auth-service/routes"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	// Initialize database
	config.InitDB()

	// Auto-migrate models
	config.DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.UserRole{},
	)

	// Seed default roles
	models.SeedRoles(config.DB)

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	routes.RegisterAuthRoutes(e)
	routes.RegisterUserRoutes(e)
	routes.RegisterRoleRoutes(e)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
			"service": "auth-service",
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Auth Service starting on port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatal(err)
	}
}
