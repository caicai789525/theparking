// internal/middleware/auth.go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"modules/config"
	"net/http"
	"strings"
)

// 示例 JWT 中间件
func JWTAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			// 这里应该返回 401 错误，而非 404
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未提供认证头"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWT.SecretKey), nil
		})
		// 处理解析错误
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "令牌解析失败: " + err.Error()})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx.Set("userID", uint(claims["user_id"].(float64)))
			ctx.Next()
		} else {
			// 这里应该返回 401 错误，而非 404
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
		}
	}
}
