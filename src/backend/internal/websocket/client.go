// Package websocket 提供 WebSocket 实时通信功能
// Client 方法实现（类型定义在 hub.go）
package websocket

import (
	"encoding/json"
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/pkg/logger"
	"github.com/gorilla/websocket"
)

// ReadPump 启动读消息 goroutine
// 持续从 WebSocket 读取消息，处理客户端消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		_ = c.Conn.Close()
	}()

	// 设置读超时
	c.Conn.SetReadDeadline(time.Now().Add(HeartbeatTimeout + HeartbeatInterval))

	// 配置 pong 处理器（客户端收到 ping 后自动回复 pong，SetPongHandler 延长读超时）
	c.Conn.SetPongHandler(func(appData string) error {
		c.Conn.SetReadDeadline(time.Now().Add(HeartbeatTimeout + HeartbeatInterval))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Warn("WS read error: %v", err)
			}
			break
		}

		// 解析客户端消息
		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Warn("WS: invalid message from user %d: %v", c.UserID, err)
			continue
		}

		// 处理各类客户端消息
		switch msg.Type {
		case "pong":
			// 客户端响应心跳，仅延长读超时即可（pong handler 已处理）
			logger.Debug("WS: received pong from user %d", c.UserID)
		case "subscribe":
			logger.Info("WS: user %d subscribed: %v", c.UserID, msg.Payload)
		case "unsubscribe":
			logger.Info("WS: user %d unsubscribed: %v", c.UserID, msg.Payload)
		default:
			logger.Warn("WS: unknown message type '%s' from user %d", msg.Type, c.UserID)
		}
	}
}

// WritePump 启动写消息 goroutine
// 持续从 Send 通道读取消息并写入 WebSocket
func (c *Client) WritePump() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			// 设置写超时
			c.Conn.SetWriteDeadline(time.Now().Add(PushDeadline))

			if !ok {
				// Send 通道已关闭，发送 WebSocket 关闭帧
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 写入消息
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Error("WS: get writer failed for user %d: %v", c.UserID, err)
				return
			}
			_, err = w.Write(message)
			if err != nil {
				logger.Error("WS: write failed for user %d: %v", c.UserID, err)
				return
			}

			// 批量合并通道中积压的消息，减少网络开销
			n := len(c.Send)
			for i := 0; i < n; i++ {
				queued, ok := <-c.Send
				if !ok {
					break
				}
				// 在消息后追加换行，再写入
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(queued)
			}

			if err := w.Close(); err != nil {
				logger.Error("WS: close writer failed for user %d: %v", c.UserID, err)
				return
			}

		case <-ticker.C:
			// 发送心跳 ping
			c.Conn.SetWriteDeadline(time.Now().Add(PushDeadline))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error("WS: heartbeat ping failed for user %d: %v", c.UserID, err)
				return
			}
			logger.Debug("WS: sent ping to user %d", c.UserID)
		}
	}
}

// SendJSON 线程安全地发送 JSON 消息（供外部调用）
func (c *Client) SendJSON(event string, payload interface{}) error {
	msg := newMessage(event, payload)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.Send <- data:
		return nil
	case <-time.After(PushDeadline):
		return nil
	}
}
