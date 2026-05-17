// Package handler 提供 HTTP 请求处理函数
// 按业务模块划分：health, auth, user, strategy, order, portfolio, backtest
package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/internal/config"
	"github.com/Uncle0206061/zeroquant2/backend/internal/websocket"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// startTime 服务启动时间
var startTime = time.Now()

// HealthCheck 健康检查
// GET /api/v1/health
// @Summary 健康检查
// @Tags 系统
// @Produce json
// @Success 200 {object} response.Response
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	response.Success(c, gin.H{
		"status":    "ok",
		"service":  "zeroquant-backend",
		"version":  "1.0.0",
		"uptime_s": int(time.Since(startTime).Seconds()),
	})
}

// Ping 心跳检查（简化版）
// GET /api/v1/ping
// @Summary 心跳检查
// @Tags 系统
// @Produce json
// @Success 200 {object} response.Response
// @Router /ping [get]
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "pong",
	})
}

// StatsHandler 系统监控指标
// GET /api/v1/stats
// 返回 CPU、内存、DB连接池、Redis连接、WebSocket连接数
// @Summary 系统监控指标
// @Tags 系统
// @Produce json
// @Success 200 {object} response.Response
// @Router /stats [get]
func StatsHandler(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	stats := gin.H{
		"uptime_seconds": int(time.Since(startTime).Seconds()),
		"memory": gin.H{
			"alloc_mb":   memStats.Alloc / 1024 / 1024,
			"sys_mb":     memStats.Sys / 1024 / 1024,
			"num_gc":     memStats.NumGC,
			"goroutines": runtime.NumGoroutine(),
		},
		"websocket": gin.H{},
	}

	// WebSocket 统计
	if hub := websocket.GetHub(); hub != nil {
		stats["websocket"] = hub.Stats()
	}

	// 数据库连接池统计
	if db := config.GetDB(); db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			dbStats := sqlDB.Stats()
			stats["database"] = gin.H{
				"max_open_connections": dbStats.MaxOpenConnections,
				"open_connections":     dbStats.OpenConnections,
				"in_use":              dbStats.InUse,
				"idle":                dbStats.Idle,
				"wait_count":          dbStats.WaitCount,
				"wait_duration_ms":    dbStats.WaitDuration.Milliseconds(),
			}
		}
	}

	// Redis 连接池统计
	if rdb := config.GetRedis(); rdb != nil {
		poolStats := rdb.PoolStats()
		stats["redis"] = gin.H{
			"total_connections": poolStats.TotalConns,
			"idle_connections":  poolStats.IdleConns,
			"stale_connections": poolStats.StaleConns,
		}
	}

	response.Success(c, stats)
}