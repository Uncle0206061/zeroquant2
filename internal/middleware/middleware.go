// Package middleware 提供 Gin 中间件
// 包含：CORS, JWT 认证, 结构化日志(request_id), Recovery, 限流
package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/internal/websocket"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/jwt"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/logger"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// RequestID 中间件 - 为每个请求生成唯一 request_id
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()[:8] // 短ID，便于日志追踪
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// JWTAuth 中间件 - JWT 认证
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetString("request_id")

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims, err := jwt.ValidateToken(tokenString)
		if err != nil {
			logger.WarnR(requestID, "JWT validation failed: %v", err)
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// StructuredLogger 中间件 - 结构化请求日志（含 request_id）
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetString("request_id")
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		cost := time.Since(start)
		status := c.Writer.Status()

		// 根据状态码选择日志级别
		switch {
		case status >= 500:
			logger.ErrorR(requestID, "%s %s %d %v %s",
				method, path, status, cost, c.ClientIP())
		case status >= 400:
			logger.WarnR(requestID, "%s %s %d %v %s",
				method, path, status, cost, c.ClientIP())
		default:
			logger.InfoR(requestID, "%s %s %d %v %s",
				method, path, status, cost, c.ClientIP())
		}
	}
}

// Recovery 中间件 - 错误恢复 + ERROR 告警推送
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := c.GetString("request_id")

				// 记录堆栈
				stack := string(debug.Stack())
				logger.ErrorR(requestID, "Panic recovered: %v\n%s", err, stack)

				// 向管理员 WebSocket 推送系统异常告警
				pushAlertToAdmins(fmt.Sprintf("Server panic: %v", err))

				// 返回 500
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    50001,
					"message": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

// RateLimit 中间件 - 简易限流（单IP 60次/分钟）
func RateLimit() gin.HandlerFunc {
	// 使用 map 做简易内存限流（生产环境应替换为 Redis 限流）
	type visitor struct {
		count    int
		expiryAt time.Time
	}
	visitors := make(map[string]*visitor)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		now := time.Now()
		if v, ok := visitors[ip]; ok {
			if now.After(v.expiryAt) {
				// 窗口过期，重置
				v.count = 1
				v.expiryAt = now.Add(time.Minute)
			} else {
				v.count++
				if v.count > 60 {
					requestID := c.GetString("request_id")
					logger.WarnR(requestID, "Rate limit exceeded for IP: %s", ip)
					response.Fail(c, 42901, "请求过于频繁，请稍后再试")
					c.Abort()
					return
				}
			}
		} else {
			visitors[ip] = &visitor{count: 1, expiryAt: now.Add(time.Minute)}
		}

		c.Next()
	}
}

// pushAlertToAdmins 向管理员推送系统异常告警
func pushAlertToAdmins(message string) {
	if hub := websocket.GetHub(); hub != nil {
		hub.BroadcastToAdmins(websocket.EventSystemAlert, map[string]interface{}{
			"level":   "error",
			"message": message,
			"time":    time.Now().Format(time.RFC3339),
		})
	}
}
