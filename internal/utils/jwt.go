package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义 JWT 声明
type Claims struct {
	UserID   uint     `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成 JWT 令牌
func GenerateJWT(secret string, userID uint, username string, roles []string, expiresIn time.Duration) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseJWT 解析 JWT 令牌
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

	return nil, errors.New("invalid token claims")
}

var jwtSecret []byte = []byte("secretkey") // 需改为配置动态加载

//func InitJWT(secret string) {
//	jwtSecret = []byte(secret)
//}

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
