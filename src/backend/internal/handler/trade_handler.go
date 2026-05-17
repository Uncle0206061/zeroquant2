// Package handler 提供 HTTP 处理器
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Uncle0206061/zeroquant2/backend/internal/service"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"
)

// TradeHandler 交易处理器（订单/持仓/账户）
type TradeHandler struct {
	orderService *service.OrderService
}

// NewTradeHandler 创建 TradeHandler
func NewTradeHandler(orderService *service.OrderService) *TradeHandler {
	return &TradeHandler{orderService: orderService}
}

// CreateSimAccount POST /api/v1/account/simulate/create
// @Summary 创建模拟账户
// @Tags 模拟交易
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /account/simulate/create [post]
func (h *TradeHandler) CreateSimAccount(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	portfolio, err := h.orderService.CreateSimAccount(userID.(int64))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"account_id": portfolio.ID,
		"user_id":    portfolio.UserID,
		"balance":    portfolio.Balance,
	})
}

// GetAccount GET /api/v1/account
// @Summary 查询模拟账户
// @Tags 模拟交易
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /account [get]
func (h *TradeHandler) GetAccount(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	portfolio, err := h.orderService.GetAccount(userID.(int64))
	if err != nil {
		response.InvalidParam(c, "模拟账户不存在，请先创建账户")
		return
	}

	response.Success(c, portfolio)
}

// SubmitOrder POST /api/v1/order/submit
// @Summary 提交订单
// @Tags 模拟交易
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "订单参数: stock_code, direction(1买入/2卖出), quantity, price(0市价)"
// @Success 200 {object} response.Response
// @Router /order/submit [post]
func (h *TradeHandler) SubmitOrder(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	var req struct {
		TaskID    string  `json:"task_id"`
		StockCode string  `json:"stock_code" binding:"required"`
		StockName string  `json:"stock_name"`
		Direction int8    `json:"direction" binding:"required"` // 1=买入 2=卖出
		Price     float64 `json:"price"`                        // 0=市价
		Quantity  int     `json:"quantity" binding:"required"`
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

	order, err := h.orderService.SubmitOrder(
		userID.(int64),
		req.TaskID,
		req.StockCode,
		req.StockName,
		req.Direction,
		req.Price,
		req.Quantity,
	)

	if err != nil {
		if order != nil && order.Status == "rejected" {
			response.InvalidParam(c, order.RejectReason)
			return
		}
		response.ServerError(c, err.Error())
		return
	}

	if order.Status == "rejected" {
		response.InvalidParam(c, order.RejectReason)
		return
	}

	response.Success(c, order)
}

// ListOrders GET /api/v1/order/list
// @Summary 订单列表
// @Tags 模拟交易
// @Produce json
// @Security BearerAuth
// @Param status query string false "订单状态"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /order/list [get]
func (h *TradeHandler) ListOrders(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	orders, total, err := h.orderService.ListOrders(userID.(int64), status, page, pageSize)
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

// GetOrder GET /api/v1/order/:id
// @Summary 订单详情
// @Tags 模拟交易
// @Produce json
// @Security BearerAuth
// @Param id path string true "订单ID"
// @Success 200 {object} response.Response
// @Router /order/{id} [get]
func (h *TradeHandler) GetOrder(c *gin.Context) {
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

	order, err := h.orderService.GetOrder(userID.(int64), orderID)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.Success(c, order)
}

// CancelOrder DELETE /api/v1/order/:id
// @Summary 撤单
// @Tags 模拟交易
// @Produce json
// @Security BearerAuth
// @Param id path string true "订单ID"
// @Success 200 {object} response.Response
// @Router /order/{id} [delete]
func (h *TradeHandler) CancelOrder(c *gin.Context) {
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

	err := h.orderService.CancelOrder(userID.(int64), orderID)
	if err != nil {
		response.InvalidParam(c, err.Error())
		return
	}

	response.SuccessMsg(c, "撤单成功")
}

// ListPositions GET /api/v1/position
// @Summary 持仓列表
// @Tags 模拟交易
// @Produce json
// @Security BearerAuth
// @Param stock_code query string false "股票代码"
// @Success 200 {object} response.Response
// @Router /position [get]
func (h *TradeHandler) ListPositions(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	stockCode := c.Query("stock_code")

	positions, err := h.orderService.ListPositions(userID.(int64), stockCode)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"list": positions})
}

// GetPosition GET /api/v1/position/:stock_code
// @Summary 持仓详情
// @Tags 模拟交易
// @Produce json
// @Security BearerAuth
// @Param stock_code path string true "股票代码"
// @Success 200 {object} response.Response
// @Router /position/{stock_code} [get]
func (h *TradeHandler) GetPosition(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	stockCode := c.Param("stock_code")
	if stockCode == "" {
		response.InvalidParam(c, "股票代码不能为空")
		return
	}

	positions, err := h.orderService.ListPositions(userID.(int64), stockCode)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	if len(positions) == 0 {
		response.InvalidParam(c, "持仓不存在")
		return
	}

	response.Success(c, positions[0])
}

// RegisterRoutes 注册交易相关路由
func (h *TradeHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/account/simulate/create", h.CreateSimAccount)
	rg.GET("/account", h.GetAccount)

	rg.POST("/order/submit", h.SubmitOrder)
	rg.GET("/order/list", h.ListOrders)
	rg.GET("/order/:id", h.GetOrder)
	rg.DELETE("/order/:id", h.CancelOrder)

	rg.GET("/position", h.ListPositions)
	rg.GET("/position/:stock_code", h.GetPosition)
}
