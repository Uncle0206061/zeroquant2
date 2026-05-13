// Package handler 提供 HTTP 处理器层
package handler

import (
	"github.com/Uncle0206061/zeroquant2/backend/internal/websocket"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// WSPushReq WebSocket 推送请求结构（供 Python 数据服务调用）
type WSPushReq struct {
	UserID  int64         `json:"user_id" binding:"required"`  // 推送目标用户 ID
	Event   string        `json:"event" binding:"required"`     // 事件类型
	Payload interface{}   `json:"payload"`                     // 推送数据
}

// WSPushHandler 处理来自 Python 数据服务的推送请求
// POST /api/v1/ws/push
// 注意：此接口无需 JWT 鉴权（仅内网 Python 服务调用），生产环境需限制来源 IP 或使用内网密钥
func WSPushHandler(c *gin.Context) {
	var req WSPushReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "参数错误: "+err.Error())
		return
	}

	// 参数校验
	if req.UserID <= 0 {
		response.InvalidParam(c, "user_id 必须大于 0")
		return
	}
	if req.Event == "" {
		response.InvalidParam(c, "event 不能为空")
		return
	}

	// 根据事件类型推送
	hub := websocket.GetHub()
	if hub == nil {
		response.ServerError(c, "WebSocket Hub 未初始化")
		return
	}

	err := hub.BroadcastToUser(req.UserID, req.Event, req.Payload)
	if err != nil {
		response.ServerError(c, "推送失败: "+err.Error())
		return
	}

	response.SuccessMsg(c, "推送成功")
}

// WSStatsHandler 返回 WebSocket 连接统计（仅管理员）
// GET /api/v1/ws/stats
func WSStatsHandler(c *gin.Context) {
	hub := websocket.GetHub()
	if hub == nil {
		response.ServerError(c, "WebSocket Hub 未初始化")
		return
	}

	response.Success(c, gin.H{
		"client_count": hub.ClientCount(),
	})
}
