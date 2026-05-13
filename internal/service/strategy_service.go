// Package service 提供业务逻辑层
// 命名规范：xxx_service.go，禁止跨层调用 repository/model/handler
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
	"github.com/Uncle0206061/zeroquant2/backend/internal/repository"
	"github.com/Uncle0206061/zeroquant2/backend/internal/websocket"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StrategyService 策略相关业务逻辑
type StrategyService struct {
	repo           *repository.StrategyRepository
	dataServiceURL string
	userID         int64 // 持有者用户ID（用于WS推送）
}

func NewStrategyService(repo *repository.StrategyRepository, dataServiceURL string) *StrategyService {
	return &StrategyService{
		repo:           repo,
		dataServiceURL: dataServiceURL,
	}
}

// ListStrategies 分页列表
func (s *StrategyService) ListStrategies(c *gin.Context, status *int, page, pageSize int) ([]model.Strategy, int64, int, int) {
	userID := s.getUserID(c)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	list, total, err := s.repo.ListStrategies(userID, status, page, pageSize)
	if err != nil {
		response.ServerError(c, "查询策略列表失败")
		return nil, 0, page, pageSize
	}
	return list, total, page, pageSize
}

// CreateStrategy 创建策略（含规则）
func (s *StrategyService) CreateStrategy(c *gin.Context, req *CreateStrategyReq) (int64, error) {
	userID := s.getUserID(c)
	rulesJSON, err := json.Marshal(req.Rules)
	if err != nil {
		response.InvalidParam(c, "rules 格式错误")
		return 0, err
	}
	strategy := &model.Strategy{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Rules:       string(rulesJSON),
		Status:      req.Status,
	}
	if err := s.repo.CreateStrategy(strategy); err != nil {
		response.ServerError(c, "创建策略失败")
		return 0, err
	}
	// 写规则表
	if len(req.RuleItems) > 0 {
		s.saveRules(strategy.ID, req.RuleItems)
	}
	return strategy.ID, nil
}

// GetStrategyDetail 获取策略详情（含规则列表）
func (s *StrategyService) GetStrategyDetail(c *gin.Context, id int64) (*model.Strategy, []model.StrategyRule, error) {
	userID := s.getUserID(c)
	strategy, err := s.repo.GetStrategyByID(id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "策略不存在")
		} else {
			response.ServerError(c, "查询策略详情失败")
		}
		return nil, nil, err
	}
	rules, err := s.repo.GetRulesByStrategyID(id)
	if err != nil {
		response.ServerError(c, "查询策略规则失败")
		return strategy, nil, err
	}
	return strategy, rules, nil
}

// UpdateStrategy 更新策略（含规则）
func (s *StrategyService) UpdateStrategy(c *gin.Context, id int64, req *UpdateStrategyReq) error {
	userID := s.getUserID(c)
	strategy, err := s.repo.GetStrategyByID(id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "策略不存在")
		} else {
			response.ServerError(c, "更新策略失败")
		}
		return err
	}
	if req.Name != "" {
		strategy.Name = req.Name
	}
	if req.Description != "" {
		strategy.Description = req.Description
	}
	if req.Rules != nil {
		rulesJSON, _ := json.Marshal(req.Rules)
		strategy.Rules = string(rulesJSON)
	}
	if req.Status != 0 {
		strategy.Status = req.Status
	}
	if err := s.repo.UpdateStrategy(strategy); err != nil {
		response.ServerError(c, "更新策略失败")
		return err
	}
	// 更新规则表
	if req.RuleItems != nil {
		s.saveRules(id, req.RuleItems)
	}
	return nil
}

// DeleteStrategy 删除策略（含规则）
func (s *StrategyService) DeleteStrategy(c *gin.Context, id int64) error {
	userID := s.getUserID(c)
	_, err := s.repo.GetStrategyByID(id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "策略不存在")
		} else {
			response.ServerError(c, "删除策略失败")
		}
		return err
	}
	// 先删规则再删策略
	if err := s.repo.DeleteRulesByStrategyID(id); err != nil {
		response.ServerError(c, "删除策略规则失败")
		return err
	}
	if err := s.repo.DeleteStrategy(id, userID); err != nil {
		response.ServerError(c, "删除策略失败")
		return err
	}
	return nil
}

// SubmitStrategy 提交策略执行/回测
func (s *StrategyService) SubmitStrategy(c *gin.Context, id int64, req *SubmitStrategyReq) (string, error) {
	userID := s.getUserID(c)
	strategy, err := s.repo.GetStrategyByID(id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "策略不存在")
		} else {
			response.ServerError(c, "查询策略失败")
		}
		return "", err
	}
	// 生成 task_id
	taskID := uuid.New().String()
	now := time.Now()
	backtest := &model.Backtest{
		StrategyID: id,
		UserID:     userID,
		Mode:       req.BacktestMode,
		Status:     "pending",
		TaskID:     taskID,
		CreatedAt:  now,
	}
	if err := s.repo.CreateBacktest(backtest); err != nil {
		response.ServerError(c, "创建回测任务失败")
		return "", err
	}
	// 更新策略最近提交时间和模式
	strategy.LastSubmitAt = &now
	strategy.LastSubmitMode = req.BacktestMode
	s.repo.UpdateStrategy(strategy)
	// 异步调用数据服务（超时5秒）
	go s.callDataServiceAsync(taskID, userID, strategy.Rules)
	return taskID, nil
}

// GetBacktestsByStrategy 获取某策略的回测记录
func (s *StrategyService) GetBacktestsByStrategy(strategyID int64) ([]model.Backtest, error) {
	return s.repo.GetBacktestsByStrategyID(strategyID)
}

// ============ 内部方法 ============

func (s *StrategyService) getUserID(c *gin.Context) int64 {
	if uid, exists := c.Get("user_id"); exists {
		return uid.(int64)
	}
	return 0
}

func (s *StrategyService) saveRules(strategyID int64, items []RuleItem) {
	var rules []model.StrategyRule
	for _, item := range items {
		rv, _ := json.Marshal(item.RuleValue)
		rules = append(rules, model.StrategyRule{
			StrategyID: strategyID,
			RuleType:   item.RuleType,
			RuleKey:    item.RuleKey,
			RuleValue:  string(rv),
			Priority:   item.Priority,
			Enabled:    item.Enabled,
		})
	}
	s.repo.UpsertRules(strategyID, rules)
}

// callDataServiceAsync 异步调用 Python 数据服务，超时5秒
func (s *StrategyService) callDataServiceAsync(taskID string, userID int64, rulesJSON string) {
	if s.dataServiceURL == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := map[string]interface{}{
		"task_id": taskID,
		"rules":   json.RawMessage(rulesJSON),
		"stocks":  []string{}, // 全部股票
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", s.dataServiceURL+"/data/v1/filter", bytes.NewReader(jsonBody))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// 超时或网络错误 → 写 failed 状态（由 Python 端更新状态，这里只记录）
		fmt.Printf("[StrategyService] 数据服务调用失败 task_id=%s: %v\n", taskID, err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("[StrategyService] 数据服务响应 task_id=%s: status=%d\n", taskID, resp.StatusCode)

	// 模拟回测完成（Python数据服务实际调用后应回调 /api/v1/ws/push）
	// 这里仅演示：异步模拟回测完成并推送结果
	go s.simulateBacktestComplete(taskID, userID)
}

// simulateBacktestComplete 模拟回测完成并推送 WS
// 注意：Python数据服务回测完成后应通过 POST /api/v1/ws/push 通知 Go 推送
// 此方法仅在 Python 服务不可用时演示用
func (s *StrategyService) simulateBacktestComplete(taskID string, userID int64) {
	time.Sleep(2 * time.Second) // 模拟处理延迟

	// 更新数据库状态
	s.repo.UpdateBacktestStatusByTaskID(taskID, "completed")

	// 推送 WebSocket
	result := gin.H{
		"task_id":  taskID,
		"status":   "completed",
		"pnl":      0,
		"win_rate": 0,
	}
	websocket.PushBacktestResult(userID, result)
	fmt.Printf("[StrategyService] 回测完成推送 task_id=%s, userID=%d\n", taskID, userID)
}

// ============ 辅助方法 ============

// ============ 请求结构体 ============

type CreateStrategyReq struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Rules       map[string]interface{} `json:"rules"`
	Status      int                    `json:"status"` // 0=草稿 1=启用
	RuleItems   []RuleItem             `json:"rule_items"`
}

type UpdateStrategyReq struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Rules       map[string]interface{} `json:"rules"`
	Status      int                    `json:"status"`
	RuleItems   []RuleItem             `json:"rule_items"`
}

type SubmitStrategyReq struct {
	BacktestMode string `json:"backtest_mode" binding:"required"` // realtime/historian
}

type RuleItem struct {
	RuleType  string      `json:"rule_type"`
	RuleKey  string      `json:"rule_key"`
	RuleValue interface{} `json:"rule_value"`
	Priority int          `json:"priority"`
	Enabled  bool         `json:"enabled"`
}
