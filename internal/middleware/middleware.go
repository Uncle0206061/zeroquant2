// Package middleware 提供 Gin 中间件（CORS, JWT 认证, 日志）
package middleware

import (
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/pkg/jwt"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/logger"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// CORS 中间件 - 跨域资源共享
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// JWTAuth 中间件 - JWT 认证
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 Token
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		// 去掉 "Bearer " 前缀
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// 验证 Token
		claims, err := jwt.ValidateToken(tokenString)
		if err != nil {
			logger.Warn("JWT validation failed: %v", err)
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 将用户信息存入 Context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// Logger 中间件 - 请求日志
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		cost := time.Since(start)
		logger.Info("| %d | %s | %s | %s | %v",
			c.Writer.Status(),
			c.Request.Method,
			path,
			c.ClientIP(),
			cost,
		)
	}
}

// Recovery 中间件 - 错误恢复
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}