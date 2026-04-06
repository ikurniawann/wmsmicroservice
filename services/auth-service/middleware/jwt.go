package middleware

import (
	"net/http"
	"strings"

	"github.com/ikurniawann/wmsmicroservice/services/auth-service/utils"
	"github.com/labstack/echo/v4"
)

// JWTMiddleware validates JWT token
func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get Authorization header
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Authorization header required",
			})
		}

		// Extract token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Validate token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid or expired token",
			})
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)

		return next(c)
	}
}

// RoleMiddleware checks if user has required role
func RoleMiddleware(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRoles, ok := c.Get("roles").([]string)
			if !ok {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Access denied",
				})
			}

			// Check if user has any of the required roles
			for _, requiredRole := range roles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						return next(c)
					}
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Insufficient permissions",
			})
		}
	}
}
