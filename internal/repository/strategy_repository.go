// Package repository 提供数据访问层
// 命名规范：xxx_repository.go，对应 biz_xxx 表
package repository

import (
	"gorm.io/gorm"
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
)

// StrategyRepository 策略相关表的 CRUD
type StrategyRepository struct {
	db *gorm.DB
}

func NewStrategyRepository(db *gorm.DB) *StrategyRepository {
	return &StrategyRepository{db: db}
}

// ============ biz_strategy CRUD ============

// CreateStrategy 创建策略
func (r *StrategyRepository) CreateStrategy(s *model.Strategy) error {
	return r.db.Create(s).Error
}

// GetStrategyByID 根据ID获取策略（可指定user_id过滤）
func (r *StrategyRepository) GetStrategyByID(id, userID int64) (*model.Strategy, error) {
	var s model.Strategy
	query := r.db.Where("id = ?", id)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	err := query.First(&s).Error
	return &s, err
}

// ListStrategies 分页查询策略列表
func (r *StrategyRepository) ListStrategies(userID int64, status *int, page, pageSize int) ([]model.Strategy, int64, error) {
	var list []model.Strategy
	var total int64

	query := r.db.Model(&model.Strategy{}).Where("user_id = ?", userID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// UpdateStrategy 更新策略
func (r *StrategyRepository) UpdateStrategy(s *model.Strategy) error {
	return r.db.Save(s).Error
}

// DeleteStrategy 删除策略（软删除改为直接删除）
func (r *StrategyRepository) DeleteStrategy(id, userID int64) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Strategy{}).Error
}

// ============ biz_strategy_rule CRUD ============

// CreateRule 创建规则
func (r *StrategyRepository) CreateRule(rule *model.StrategyRule) error {
	return r.db.Create(rule).Error
}

// GetRulesByStrategyID 获取某策略的全部规则
func (r *StrategyRepository) GetRulesByStrategyID(strategyID int64) ([]model.StrategyRule, error) {
	var rules []model.StrategyRule
	err := r.db.Where("strategy_id = ?", strategyID).Order("priority ASC, id ASC").Find(&rules).Error
	return rules, err
}

// DeleteRulesByStrategyID 删除某策略的全部规则
func (r *StrategyRepository) DeleteRulesByStrategyID(strategyID int64) error {
	return r.db.Where("strategy_id = ?", strategyID).Delete(&model.StrategyRule{}).Error
}

// UpsertRules 先删后插，批量保存规则
func (r *StrategyRepository) UpsertRules(strategyID int64, rules []model.StrategyRule) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("strategy_id = ?", strategyID).Delete(&model.StrategyRule{}).Error; err != nil {
			return err
		}
		if len(rules) == 0 {
			return nil
		}
		for i := range rules {
			rules[i].StrategyID = strategyID
		}
		return tx.Create(&rules).Error
	})
}

// ============ biz_backtest CRUD ============

// CreateBacktest 创建回测/执行任务记录
func (r *StrategyRepository) CreateBacktest(b *model.Backtest) error {
	return r.db.Create(b).Error
}

// GetBacktestByTaskID 根据 task_id 查询任务
func (r *StrategyRepository) GetBacktestByTaskID(taskID string) (*model.Backtest, error) {
	var b model.Backtest
	err := r.db.Where("task_id = ?", taskID).First(&b).Error
	return &b, err
}

// GetBacktestsByStrategyID 获取某策略的全部回测记录
func (r *StrategyRepository) GetBacktestsByStrategyID(strategyID int64) ([]model.Backtest, error) {
	var list []model.Backtest
	err := r.db.Where("strategy_id = ?", strategyID).Order("created_at DESC").Find(&list).Error
	return list, err
}

// UpdateBacktestStatus 更新回测状态（按 id）
func (r *StrategyRepository) UpdateBacktestStatus(id int64, status string, resultSummary string, startedAt, completedAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if resultSummary != "" {
		updates["result_summary"] = resultSummary
	}
	if startedAt != nil {
		updates["started_at"] = startedAt
	}
	if completedAt != nil {
		updates["completed_at"] = completedAt
	}
	return r.db.Model(&model.Backtest{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateBacktestStatusByTaskID 按 task_id 更新回测状态
func (r *StrategyRepository) UpdateBacktestStatusByTaskID(taskID string, status string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":        status,
		"completed_at": now,
	}
	return r.db.Model(&model.Backtest{}).Where("task_id = ?", taskID).Updates(updates).Error
}
