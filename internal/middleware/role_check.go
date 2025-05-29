// internal/middleware/role_check.go (新建文件)
package middleware

import (
	"modules/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RoleCheck(requiredRole models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权限访问"})
			return
		}

		hasRole := false
		for _, role := range userRoles.([]models.Role) {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			return
		}
		c.Next()
	}
}
