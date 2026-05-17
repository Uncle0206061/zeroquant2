// Package middleware 提供实盘模式切换中间件
// 根据 REAL_TRADE 环境变量控制实盘路由的访问权限
package middleware

import (
	"os"
	"strings"

	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// TradeModeMiddleware 实盘模式切换中间件
// REAL_TRADE=true 时允许访问 /api/v1/trade/real/* 路由
// REAL_TRADE=false 时拒绝访问实盘路由
func TradeModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		realTrade := os.Getenv("REAL_TRADE") == "true"

		// 如果访问实盘路由但实盘模式未开启，拒绝
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/trade/real") && !realTrade {
			response.Forbidden(c, "实盘模式未开启，请设置环境变量 REAL_TRADE=true")
			c.Abort()
			return
		}

		// 将实盘模式标记写入上下文
		c.Set("real_trade", realTrade)
		c.Next()
	}
}

// IsRealTradeMode 检查当前是否为实盘模式（供 service 层调用）
func IsRealTradeMode() bool {
	return os.Getenv("REAL_TRADE") == "true"
}
