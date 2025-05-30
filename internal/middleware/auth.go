package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"modules/config"
	"modules/internal/utils"
	"net/http"
	"strings"
)

// JWTAuthMiddleware JWT 认证中间件
func JWTAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			fmt.Println("未提供认证令牌，请求被拦截")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ParseJWT(tokenString, cfg.JWT.Secret)
		if err != nil {
			fmt.Printf("解析令牌失败: %v，请求被拦截\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}
