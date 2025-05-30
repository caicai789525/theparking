package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
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
			log.Println("缺少认证令牌")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
			c.Abort()
			return
		}

		log.Printf("接收到的 Authorization 头信息: %s", authHeader)
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			log.Println("无效的认证令牌格式")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌格式"})
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			log.Printf("令牌验证失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}

		log.Printf("令牌验证成功，用户 ID: %d, 用户名: %s", claims.UserID, claims.Username)
		// 将 claims 存入上下文，供后续处理使用
		c.Set("claims", claims)
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}
