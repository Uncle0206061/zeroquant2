// Package handler 提供 HTTP 请求处理层
// 命名规范：xxx_handler.go，调用 service 层，禁止跨层调用 repository/model
package handler

import (
	"strconv"

	"github.com/Uncle0206061/zeroquant2/backend/internal/service"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// StrategyHandler 策略管理 HTTP 处理
type StrategyHandler struct {
	svc *service.StrategyService
}

func NewStrategyHandler(svc *service.StrategyService) *StrategyHandler {
	return &StrategyHandler{svc: svc}
}

// ListStrategy GET /api/v1/strategy/list
// Query: page=1&page_size=20&status=1
func (h *StrategyHandler) ListStrategy(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var status *int
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			status = &v
		}
	}
	list, total, pg, ps := h.svc.ListStrategies(c, status, page, pageSize)
	response.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      pg,
		"page_size": ps,
	})
}

// CreateStrategy POST /api/v1/strategy/create
func (h *StrategyHandler) CreateStrategy(c *gin.Context) {
	var req service.CreateStrategyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "参数错误")
		return
	}
	id, err := h.svc.CreateStrategy(c, &req)
	if err != nil {
		return // 错误已在 service 层处理
	}
	response.Success(c, gin.H{"id": id, "name": req.Name})
}

// GetStrategy GET /api/v1/strategy/:id
func (h *StrategyHandler) GetStrategy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的策略ID")
		return
	}
	strategy, rules, err := h.svc.GetStrategyDetail(c, id)
	if err != nil {
		return
	}
	response.Success(c, gin.H{
		"strategy": strategy,
		"rules":    rules,
	})
}

// UpdateStrategy PUT /api/v1/strategy/:id
func (h *StrategyHandler) UpdateStrategy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的策略ID")
		return
	}
	var req service.UpdateStrategyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "参数错误")
		return
	}
	if err := h.svc.UpdateStrategy(c, id, &req); err != nil {
		return
	}
	response.Success(c, gin.H{"id": id})
}

// DeleteStrategy DELETE /api/v1/strategy/:id
func (h *StrategyHandler) DeleteStrategy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的策略ID")
		return
	}
	if err := h.svc.DeleteStrategy(c, id); err != nil {
		return
	}
	response.SuccessMsg(c, "删除成功")
}

// SubmitStrategy POST /api/v1/strategy/:id/submit
func (h *StrategyHandler) SubmitStrategy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的策略ID")
		return
	}
	var req service.SubmitStrategyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "参数错误")
		return
	}
	taskID, err := h.svc.SubmitStrategy(c, id, &req)
	if err != nil {
		return
	}
	response.Success(c, gin.H{
		"task_id": taskID,
		"status":  "pending",
	})
}

// GetBacktests GET /api/v1/strategy/:id/backtests
func (h *StrategyHandler) GetBacktests(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParam(c, "无效的策略ID")
		return
	}
	list, err := h.svc.GetBacktestsByStrategy(id)
	if err != nil {
		response.ServerError(c, "查询回测记录失败")
		return
	}
	response.Success(c, gin.H{"list": list})
}
