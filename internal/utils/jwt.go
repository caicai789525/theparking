// internal/utils/jwt.go
package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"modules/internal/models"
	"time"
)

var jwtSecret []byte = []byte("secretkey") // 需改为配置动态加载

//func InitJWT(secret string) {
//	jwtSecret = []byte(secret)
//}

type Claims struct {
	UserID   uint          `json:"user_id"`
	Username string        `json:"username"`
	Roles    []models.Role `json:"roles"`
	jwt.RegisteredClaims
}

// GetExpirationTime 返回 JWT 的过期时间
func (c Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.ExpiresAt, nil
}

// GetIssuedAt 返回 JWT 的签发时间
func (c Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.IssuedAt, nil
}

// GetNotBefore 返回 JWT 的生效时间
func (c Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.NotBefore, nil
}

// GetIssuer 返回 JWT 的签发者
func (c Claims) GetIssuer() (string, error) {
	return c.RegisteredClaims.Issuer, nil
}

// GetSubject 返回 JWT 的主题
func (c Claims) GetSubject() (string, error) {
	return c.RegisteredClaims.Subject, nil
}

func GenerateJWT(userID uint, username string) (string, error) {
	now := time.Now()
	expireTime := now.Add(24 * time.Hour)
	claims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    []models.Role{}, // 可根据实际情况填充
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "parking-app",
			Subject:   "user token",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseJWT(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}
