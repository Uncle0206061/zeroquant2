// Package middleware 提供 Gin 中间件
// 包含：CORS, JWT 认证, 结构化日志(request_id), Recovery, 限流
package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/internal/config"
	"github.com/Uncle0206061/zeroquant2/backend/internal/websocket"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/jwt"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/logger"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// rateLimitVisitor 内存限流用（Redis 不可用时的降级方案）
type rateLimitVisitor struct {
	count    int
	expiryAt time.Time
}

var (
	rateLimitVisitors = make(map[string]*rateLimitVisitor)
	rateLimitMu       sync.RWMutex
)

// CORS 中间件 - 跨域资源共享（从配置读取允许列表）
func CORS() gin.HandlerFunc {
	cfg := config.GetConfig()
	allowedOrigins := make(map[string]bool)
	for _, origin := range cfg.CORSAllowedOrigins {
		allowedOrigins[origin] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := false

		// 检查是否在允许列表中
		if allowedOrigins[origin] {
			allowed = true
		} else if origin == "" {
			// 无Origin头，允许（简单请求）
			allowed = true
		}

		if allowed {
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			}
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Max-Age", "86400")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

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

// RateLimit 中间件 - 基于 Redis 的分布式限流（单IP 60次/分钟）
func RateLimit() gin.HandlerFunc {
	// Redis 限流键前缀
	keyPrefix := "rate_limit:"

	// 启动后台清理goroutine（仅启动一次，使用 sync.Once）
	var cleanupOnce sync.Once

	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := keyPrefix + ip

		// 尝试使用 Redis 限流
		redisClient := config.GetRedis()
		if redisClient != nil {
			ctx := c.Request.Context()
			// 使用 INCR 实现滑动窗口限流
			pipe := redisClient.Pipeline()
			pipe.Incr(ctx, key)
			pipe.Expire(ctx, key, time.Minute)
			cmds, err := pipe.Exec(ctx)
			if err == nil && cmds[0].(*redis.IntCmd).Val() > 60 {
				requestID := c.GetString("request_id")
				logger.WarnR(requestID, "Rate limit exceeded for IP: %s", ip)
				response.Fail(c, 42901, "请求过于频繁，请稍后再试")
				c.Abort()
				return
			}
			// Redis 限流成功，放行
			c.Next()
			return
		}

		// Redis 不可用时，使用内存限流作为降级方案
		cleanupOnce.Do(func() {
			go func() {
				ticker := time.NewTicker(5 * time.Minute)
				for range ticker.C {
					now := time.Now()
					rateLimitMu.RLock()
					for ip, v := range rateLimitVisitors {
						if now.After(v.expiryAt) {
							delete(rateLimitVisitors, ip)
						}
					}
					rateLimitMu.RUnlock()
				}
			}()
		})

		rateLimitMu.Lock()
		defer rateLimitMu.Unlock()

		now := time.Now()
		if v, ok := rateLimitVisitors[ip]; ok {
			if now.After(v.expiryAt) {
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
			rateLimitVisitors[ip] = &rateLimitVisitor{count: 1, expiryAt: now.Add(time.Minute)}
		}

		c.Next()
	}
}

// pushAlertToAdmins 向管理员推送系统异常告警
func pushAlertToAdmins(message string) {
	if hub := websocket.GetHub(); nil != hub {
		hub.BroadcastToAdmins(websocket.EventSystemAlert, map[string]interface{}{
			"level":   "error",
			"message": message,
			"time":    time.Now().Format(time.RFC3339),
		})
	}
}

// RequireAdmin 中间件 - 管理员权限检查
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			userID, _ := c.Get("user_id")
			logger.WarnR(c.GetString("request_id"), "Unauthorized admin access from user %v", userID)
			response.Forbidden(c, "需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
