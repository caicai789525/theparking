// internal/utils/jwt.go
package utils

import (
	"github.com/dgrijalva/jwt-go"
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
	jwt.StandardClaims
}

func GenerateJWT(userID uint, username string) (string, error) {
	now := time.Now()
	expireTime := now.Add(24 * time.Hour)
	claims := Claims{
		UserID:   userID,
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  now.Unix(),
			Issuer:    "parking-app",
			Subject:   "user token",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
