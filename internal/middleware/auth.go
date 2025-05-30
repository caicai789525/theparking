package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"modules/config"
	"modules/pkg/logger"
	"net/http"
	"strings"
)

func JWTAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Println("Entering JWTAuthMiddleware, Path:", ctx.Request.URL.Path)
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			fmt.Println("No Authorization header provided")
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

		if err != nil {
			logger.Log.Error("令牌解析失败", zap.String("path", ctx.Request.URL.Path), zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID, ok := claims["user_id"].(float64)
			if !ok {
				logger.Log.Error("无法从令牌中获取用户 ID", zap.String("path", ctx.Request.URL.Path))
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
				return
			}
			ctx.Set("userID", uint(userID))
			ctx.Next()
		} else {
			logger.Log.Error("无效的令牌", zap.String("path", ctx.Request.URL.Path))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
		}
	}
}
