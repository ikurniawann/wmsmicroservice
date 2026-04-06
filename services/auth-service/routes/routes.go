package routes

import (
	"github.com/ikurniawann/wmsmicroservice/services/auth-service/handlers"
	"github.com/ikurniawann/wmsmicroservice/services/auth-service/middleware"
	"github.com/labstack/echo/v4"
)

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(e *echo.Echo) {
	auth := e.Group("/api/v1/auth")
	
	auth.POST("/register", handlers.Register)
	auth.POST("/login", handlers.Login)
	auth.POST("/logout", handlers.Logout)
	auth.POST("/refresh", handlers.Refresh)
}

// RegisterUserRoutes registers user management routes (protected)
func RegisterUserRoutes(e *echo.Echo) {
	users := e.Group("/api/v1/users")
	users.Use(middleware.JWTMiddleware)
	
	users.GET("/me", handlers.Me)
}

// RegisterRoleRoutes registers role management routes (protected)
func RegisterRoleRoutes(e *echo.Echo) {
	roles := e.Group("/api/v1/roles")
	roles.Use(middleware.JWTMiddleware)
	// TODO: Add role handlers
	roles.GET("", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "Roles endpoint"})
	})
}
