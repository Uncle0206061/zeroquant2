// Package handler 提供实盘交易 HTTP 处理器
// 路由前缀：/api/v1/trade/real/
// 安全机制：TradeModeMiddleware 控制访问 + 二次确认流程
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Uncle0206061/zeroquant2/backend/internal/service"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"
)

// RealTradeHandler 实盘交易处理器
type RealTradeHandler struct {
	realOrderService *service.RealOrderService
}

// NewRealTradeHandler 创建 RealTradeHandler
func NewRealTradeHandler(realOrderService *service.RealOrderService) *RealTradeHandler {
	return &RealTradeHandler{realOrderService: realOrderService}
}

// CreateRealAccount POST /api/v1/trade/real/account/create
// @Summary 创建实盘账户
// @Tags 实盘交易
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /trade/real/account/create [post]
func (h *RealTradeHandler) CreateRealAccount(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	portfolio, err := h.realOrderService.CreateRealAccount(userID.(int64))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"account_id": portfolio.ID,
		"user_id":    portfolio.UserID,
		"balance":    portfolio.Balance,
		"mode":       "real",
	})
}

// GetRealAccount GET /api/v1/trade/real/account
// @Summary 查询实盘账户
// @Tags 实盘交易
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /trade/real/account [get]
func (h *RealTradeHandler) GetRealAccount(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	portfolio, err := h.realOrderService.GetRealAccount(userID.(int64))
	if err != nil {
		response.InvalidParam(c, "实盘账户不存在，请先创建账户")
		return
	}

	response.Success(c, portfolio)
}

// SubmitRealOrder POST /api/v1/trade/real/order/submit
// @Summary 实盘下单
// @Tags 实盘交易
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /trade/real/order/submit [post]
func (h *RealTradeHandler) SubmitRealOrder(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	var req struct {
		StockCode string  `json:"stock_code" binding:"required"`
		StockName string  `json:"stock_name"`
		Direction int8    `json:"direction" binding:"required"` // 1=买入 2=卖出
		Price     float64 `json:"price"`                        // 0=市价
		Quantity  int     `json:"quantity" binding:"required"`
		OrderType int8    `json:"order_type"` // 1=市价 2=限价，默认1
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "参数错误："+err.Error())
		return
	}

	if req.Quantity%100 != 0 {
		response.InvalidParam(c, "数量必须是100的整数倍（手）")
		return
	}

	if req.Direction != 1 && req.Direction != 2 {
		response.InvalidParam(c, "direction 必须为 1(买入) 或 2(卖出)")
		return
	}

	if req.OrderType == 0 {
		req.OrderType = 1 // 默认市价
	}

	order, err := h.realOrderService.SubmitRealOrder(
		userID.(int64),
		req.StockCode,
		req.StockName,
		req.Direction,
		req.Price,
		req.Quantity,
		req.OrderType,
	)

	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, order)
}

// ConfirmRealOrder POST /api/v1/trade/real/order/confirm
// 二次确认：pending → confirmed → 提交券商
// @Summary 实盘二次确认
// @Tags 实盘交易
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /trade/real/order/confirm [post]
func (h *RealTradeHandler) ConfirmRealOrder(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	var req struct {
		OrderID string `json:"order_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "参数错误：order_id 必填")
		return
	}

	order, err := h.realOrderService.ConfirmRealOrder(userID.(int64), req.OrderID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, order)
}

// CancelRealOrder DELETE /api/v1/trade/real/order/:id
// @Summary 实盘撤单
// @Tags 实盘交易
// @Produce json
// @Security BearerAuth
// @Param id path string true "订单ID"
// @Success 200 {object} response.Response
// @Router /trade/real/order/{id} [delete]
func (h *RealTradeHandler) CancelRealOrder(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		response.InvalidParam(c, "订单ID不能为空")
		return
	}

	err := h.realOrderService.CancelRealOrder(userID.(int64), orderID)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.SuccessMsg(c, "撤单成功")
}

// ListRealOrders GET /api/v1/trade/real/order/list
// @Summary 实盘订单列表
// @Tags 实盘交易
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /trade/real/order/list [get]
func (h *RealTradeHandler) ListRealOrders(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	orders, total, err := h.realOrderService.ListRealOrders(userID.(int64), status, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetRealOrder GET /api/v1/trade/real/order/:id
// @Summary 实盘订单详情
// @Tags 实盘交易
// @Produce json
// @Security BearerAuth
// @Param id path string true "订单ID"
// @Success 200 {object} response.Response
// @Router /trade/real/order/{id} [get]
func (h *RealTradeHandler) GetRealOrder(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		response.InvalidParam(c, "订单ID不能为空")
		return
	}

	order, err := h.realOrderService.GetRealOrder(userID.(int64), orderID)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, order)
}

// ListRealPositions GET /api/v1/trade/real/position
// @Summary 实盘持仓列表
// @Tags 实盘交易
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /trade/real/position [get]
func (h *RealTradeHandler) ListRealPositions(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	stockCode := c.Query("stock_code")

	positions, err := h.realOrderService.ListRealPositions(userID.(int64), stockCode)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"list": positions})
}

// ListRealTradeLogs GET /api/v1/trade/real/log
// @Summary 实盘操作日志
// @Tags 实盘交易
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /trade/real/log [get]
func (h *RealTradeHandler) ListRealTradeLogs(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	logs, total, err := h.realOrderService.ListRealTradeLogs(userID.(int64), page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 注册实盘交易路由
func (h *RealTradeHandler) RegisterRoutes(rg *gin.RouterGroup) {
	real := rg.Group("/trade/real")
	{
		real.POST("/account/create", h.CreateRealAccount)
		real.GET("/account", h.GetRealAccount)

		real.POST("/order/submit", h.SubmitRealOrder)
		real.POST("/order/confirm", h.ConfirmRealOrder)
		real.GET("/order/list", h.ListRealOrders)
		real.GET("/order/:id", h.GetRealOrder)
		real.DELETE("/order/:id", h.CancelRealOrder)

		real.GET("/position", h.ListRealPositions)

		real.GET("/log", h.ListRealTradeLogs)
	}
}
