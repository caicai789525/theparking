// internal/middleware/auth.go
package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"modules/config"
	utils "modules/internal/utils"
	"net/http"
	"strings"
)

// internal/middleware/auth.go
func JWTAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("进入 JWTAuthMiddleware，请求路径:", c.Request.URL.Path)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ParseJWT(tokenString, cfg.JWT.Secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}
