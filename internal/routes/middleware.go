// internal/routes/middleware.go
package routes

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"modules/internal/services"
	"modules/pkg/logger"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
			return
		}

		// 解析Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "令牌格式错误"})
			return
		}

		// 验证JWT
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("无效的签名方法")
			}
			return []byte(authService.Cfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			logger.Log.Warn("无效的JWT令牌", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效令牌"})
			return
		}

		// 设置用户信息到上下文
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("userID", claims["userID"])
		}

		c.Next()
	}
}
