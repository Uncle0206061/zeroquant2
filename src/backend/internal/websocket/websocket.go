// Package websocket 提供 WebSocket 实时通信功能
// 预置配置：心跳间隔 10 秒 / 断开重连最多 10 次间隔 2 秒 / 推送延迟 ≤200ms / 最大连接100
package websocket

import (
	"net/http"
	"strings"
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/internal/config"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/jwt"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/logger"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境：从配置读取允许的 origin
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // 无 Origin 头允许（简单请求）
		}

		cfg := config.GetConfig()
		for _, allowed := range cfg.CORSAllowedOrigins {
			if origin == allowed {
				return true
			}
		}
		// 不在允许列表中拒绝
		logger.Warn("WS: origin not allowed: %s", origin)
		return false
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
	logger.Info("WebSocket Hub initialized, max_connections=%d", MaxConnections)
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
	role := claims.Role

	// ============ 升级为 WebSocket ============
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WS: failed to upgrade: %v", err)
		return
	}

	// 创建客户端并注册到 Hub
	client := &Client{
		UserID:      userID,
		Role:        role,
		Conn:        conn,
		Hub:         globalHub,
		Send:        make(chan []byte, 256),
		ConnectedAt: time.Now(),
	}

	globalHub.Register(client)
	logger.Info("WS: client connected, userID=%d, role=%s, ip=%s", userID, role, c.ClientIP())

	// 发送欢迎消息
	_ = client.SendJSON(EventWelcome, gin.H{
		"message":      "Connected to ZeroQuant 2.0",
		"user_id":      userID,
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

// PushBacktestResult 推送回测结果
func PushBacktestResult(userID int64, payload interface{}) {
	if globalHub == nil {
		return
	}
	globalHub.BroadcastToUser(userID, EventBacktestResult, payload)
}

// PushStrategySignal 推送策略信号
func PushStrategySignal(userID int64, payload interface{}) {
	if globalHub == nil {
		return
	}
	globalHub.BroadcastToUser(userID, EventStrategySignal, payload)
}

// PushSystemAlert 推送系统异常告警（给管理员）
func PushSystemAlert(message string) {
	if globalHub == nil {
		return
	}
	globalHub.BroadcastToAdmins(EventSystemAlert, map[string]interface{}{
		"level":   "error",
		"message": message,
		"time":    time.Now().Format(time.RFC3339),
	})
}
