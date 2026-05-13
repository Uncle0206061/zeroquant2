// Package websocket 提供 WebSocket 实时通信功能
// 核心组件：Hub（连接管理器）+ Client（单个连接封装）
package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/pkg/logger"
	"github.com/gorilla/websocket"
)

// ============ 常量定义 ============

const (
	EventWelcome          = "welcome"
	EventPing             = "ping"
	EventPong             = "pong"
	EventPositionUpdate   = "position_update"
	EventOrderUpdate      = "order_update"
	EventBacktestProgress = "backtest_progress"
	EventBacktestResult   = "backtest_result"
	EventStrategySignal   = "strategy_signal"
	EventError            = "error"
)

// 推送配置（技术规范强制）
const (
	HeartbeatInterval = 10 * time.Second // 心跳间隔
	HeartbeatTimeout  = 10 * time.Second // 无响应超时
	PushDeadline      = 200 * time.Millisecond // 推送超时
	MaxReconnect      = 10               // 最大重连次数（客户端视角）
	ReconnectDelay    = 2 * time.Second  // 重连间隔
	MessageBufferSize = 100              // 每个用户消息保留条数
)

// WSMessage WebSocket 消息结构（统一格式）
type WSMessage struct {
	Type      string      `json:"type"`       // 事件类型
	Payload   interface{} `json:"payload"`    // 数据内容
	Timestamp int64       `json:"timestamp"`  // 毫秒时间戳
}

// Client 代表一个 WebSocket 客户端连接
type Client struct {
	UserID    int64            // 用户ID
	Conn      *websocket.Conn  // WebSocket 连接
	Hub       *Hub             // 所属 Hub
	Send      chan []byte       // 发送消息缓冲通道
	ConnectedAt time.Time       // 连接时间
}

// Hub 管理所有 WebSocket 客户端连接
type Hub struct {
	mu       sync.RWMutex
	clients  map[int64]*Client // userID → Client（同用户仅一个活跃连接）
	register   chan *Client    // 注册新连接
	unregister chan *Client    // 注销连接
}

// NewHub 创建新的 Hub 实例
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]*Client),
		register:   make(chan *Client, 10),
		unregister: make(chan *Client, 10),
	}
}

// Run 启动 Hub 的事件循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// 如果该用户已有连接，先踢掉旧连接
			if old, ok := h.clients[client.UserID]; ok {
				close(old.Send)
				_ = old.Conn.Close()
				logger.Info("WS: kicked old connection for user %d", client.UserID)
			}
			h.clients[client.UserID] = client
			logger.Info("WS: registered client, userID=%d, total=%d", client.UserID, len(h.clients))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if existing, ok := h.clients[client.UserID]; ok && existing == client {
				delete(h.clients, client.UserID)
				close(client.Send)
				logger.Info("WS: unregistered client, userID=%d, total=%d", client.UserID, len(h.clients))
			}
			h.mu.Unlock()
		}
	}
}

// Register 向 Hub 注册客户端（线程安全）
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 从 Hub 注销客户端（线程安全）
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// GetClient 根据 userID 查找客户端
func (h *Hub) GetClient(userID int64) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[userID]
}

// ClientCount 返回当前连接数
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// BroadcastToUser 向指定用户推送消息（200ms 超时）
func (h *Hub) BroadcastToUser(userID int64, event string, payload interface{}) error {
	msg := newMessage(event, payload)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	client := h.GetClient(userID)
	if client == nil {
		return nil // 用户不在线，忽略
	}

	select {
	case client.Send <- data:
		return nil
	case <-time.After(PushDeadline):
		logger.Warn("WS: push timeout for user %d, event=%s", userID, event)
		return nil
	}
}

// Broadcast 向所有在线用户广播消息
func (h *Hub) Broadcast(event string, payload interface{}) {
	msg := newMessage(event, payload)
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("WS: marshal broadcast failed: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for userID, client := range h.clients {
		select {
		case client.Send <- data:
		case <-time.After(PushDeadline):
			logger.Warn("WS: broadcast timeout for user %d", userID)
		}
	}
}

// newMessage 创建带时间戳的消息
func newMessage(event string, payload interface{}) WSMessage {
	return WSMessage{
		Type:      event,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}
}
