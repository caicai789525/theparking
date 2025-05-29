// internal/middleware/auth.go
package middleware

import (
	"github.com/gin-gonic/gin"
	utils "modules/internal/utils"
)

// internal/middleware/auth.go
func JWTAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "未提供认证令牌"})
			return
		}

		claims, err := utils.ParseJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "无效令牌"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("roles", claims.Roles) // 需要JWT包含roles字段
		c.Next()
	}
}
