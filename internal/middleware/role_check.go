// internal/middleware/role_check.go (新建文件)
package middleware

import (
	"modules/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RoleCheck 角色检查中间件
func RoleCheck(requiredRole models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户角色
		userRoles, exists := c.Get("roles")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未授权访问，缺少角色信息"})
			return
		}

		// 尝试将 userRoles 转换为 []string
		roleStrings, ok := userRoles.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "角色信息类型错误"})
			return
		}

		// 将 []string 转换为 []models.Role
		var roles []models.Role
		for _, roleStr := range roleStrings {
			roles = append(roles, models.Role(roleStr))
		}

		// 检查是否有需要的角色
		hasRequiredRole := false
		for _, role := range roles {
			if role == requiredRole {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "权限不足，需要特定角色"})
			return
		}

		c.Next()
	}
}
