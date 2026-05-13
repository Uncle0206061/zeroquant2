// Package handler 提供 HTTP 请求处理函数
// 按业务模块划分：health, auth, user, watchlist, strategy, order, portfolio, backtest
package handler

import (
	"net/http"

	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查
// GET /api/v1/health
func HealthCheck(c *gin.Context) {
	response.Success(c, gin.H{
		"status":   "ok",
		"service": "zeroquant-backend",
		"version": "0.1.0",
	})
}

// Ping 心跳检查（简化版）
// GET /api/v1/ping
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "pong",
	})
}