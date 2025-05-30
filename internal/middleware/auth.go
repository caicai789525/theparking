package middleware

import (
	"github.com/gin-gonic/gin"
	"modules/config"
	"modules/internal/services"
	"net/http"
	"strings"
)

// JWTAuthMiddleware JWT 认证中间件
func JWTAuthMiddleware(cfg *config.Config, authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
			c.Abort()
			return
		}

		// 打印接收到的 Authorization 头信息，用于调试
		print("Received Authorization header:", authHeader)

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌格式"})
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}

		// 将 claims 存入上下文，供后续处理使用
		c.Set("claims", claims)
		c.Next()
	}
}
