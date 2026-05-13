// Package jwt 提供 JWT Token 生成和验证功能
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims JWT 载荷结构
type Claims struct {
	UserID   int64  `json:"user_id"`   // 用户 ID
	Username string `json:"username"` // 用户名
	Role    string `json:"role"`    // 角色：admin, user
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT Token
func GenerateToken(userID int64, username, role string) (string, error) {
	cfg := config.GetConfig()
	
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWTExpire) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:   "zeroquant",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// ValidateToken 验证 JWT Token
func ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.GetConfig()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshToken 刷新 Token
func RefreshToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return GenerateToken(claims.UserID, claims.Username, claims.Role)
}