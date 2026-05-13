// Package websocket 提供 WebSocket 实时通信功能
// 预置配置：心跳间隔 10 秒 / 断开重连最多 10 次间隔 2 秒 / 推送延迟 ≤200ms
package websocket

import (
	"net/http"
	"strings"
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/pkg/jwt"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/logger"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源（生产环境应限制为前端域名）
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// 全局 Hub 实例（整个应用共享）
var globalHub *Hub

// InitHub 初始化全局 Hub（main.go 启动时调用）
func InitHub() {
	globalHub = NewHub()
	go globalHub.Run()
	logger.Info("WebSocket Hub initialized")
}

// GetHub 获取全局 Hub 实例
func GetHub() *Hub {
	return globalHub
}

// HandleWS 处理 WebSocket 连接
// WS /api/v1/ws
// Header: Authorization: Bearer {token}
func HandleWS(c *gin.Context) {
	// ============ JWT 鉴权 ============
	token := c.GetHeader("Authorization")
	if !strings.HasPrefix(token, "Bearer ") {
		logger.Warn("WS: missing or invalid Authorization header from %s", c.ClientIP())
		response.Unauthorized(c, "missing token")
		return
	}
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		response.Unauthorized(c, "missing token")
		return
	}

	claims, err := jwt.ValidateToken(token)
	if err != nil {
		logger.Warn("WS: invalid token from %s: %v", c.ClientIP(), err)
		response.Unauthorized(c, "invalid token")
		return
	}
	userID := claims.UserID

	// ============ 升级为 WebSocket ============
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WS: failed to upgrade: %v", err)
		return
	}

	// 创建客户端并注册到 Hub
	client := &Client{
		UserID:    userID,
		Conn:      conn,
		Hub:       globalHub,
		Send:      make(chan []byte, 256),
		ConnectedAt: time.Now(),
	}

	globalHub.Register(client)
	logger.Info("WS: client connected, userID=%d, ip=%s", userID, c.ClientIP())

	// 发送欢迎消息
	_ = client.SendJSON(EventWelcome, gin.H{
		"message":   "Connected to ZeroQuant 2.0",
		"user_id":   userID,
		"connected_at": client.ConnectedAt,
	})

	// 启动读写 goroutine
	go client.WritePump()
	client.ReadPump() // 阻塞直到连接断开
}

// nowMilli 返回当前毫秒时间戳
func nowMilli() int64 {
	return time.Now().UnixMilli()
}

// PushPositionUpdate 推送持仓变化（供 service 层调用）
func PushPositionUpdate(userID int64, payload interface{}) {
	if globalHub == nil {
		return
	}
	globalHub.BroadcastToUser(userID, EventPositionUpdate, payload)
}

// PushOrderUpdate 推送订单变化（供 service 层调用）
func PushOrderUpdate(userID int64, payload interface{}) {
	if globalHub == nil {
		return
	}
	globalHub.BroadcastToUser(userID, EventOrderUpdate, payload)
}

// PushBacktestProgress 推送回测进度
func PushBacktestProgress(userID int64, payload interface{}) {
	if globalHub == nil {
		return
	}
	globalHub.BroadcastToUser(userID, EventBacktestProgress, payload)
}

// PushBacktestResult 推送回测结果（供 service 层调用）
func PushBacktestResult(userID int64, payload interface{}) {
	if globalHub == nil {
		return
	}
	globalHub.BroadcastToUser(userID, EventBacktestResult, payload)
}

// PushStrategySignal 推送策略信号（供 service 层调用）
func PushStrategySignal(userID int64, payload interface{}) {
	if globalHub == nil {
		return
	}
	globalHub.BroadcastToUser(userID, EventStrategySignal, payload)
}
