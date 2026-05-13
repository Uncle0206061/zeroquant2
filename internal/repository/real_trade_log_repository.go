// Package repository 提供实盘操作日志数据访问层
package repository

import (
	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
	"gorm.io/gorm"
)

// RealTradeLogRepository 实盘操作日志数据访问
type RealTradeLogRepository struct {
	db *gorm.DB
}

// NewRealTradeLogRepository 创建 RealTradeLogRepository
func NewRealTradeLogRepository(db *gorm.DB) *RealTradeLogRepository {
	return &RealTradeLogRepository{db: db}
}

// Create 创建操作日志
func (r *RealTradeLogRepository) Create(log *model.RealTradeLog) error {
	return r.db.Create(log).Error
}

// ListByUser 查询用户操作日志（按时间倒序）
func (r *RealTradeLogRepository) ListByUser(userID int64, page, pageSize int) ([]model.RealTradeLog, int64, error) {
	var logs []model.RealTradeLog
	var total int64

	query := r.db.Model(&model.RealTradeLog{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// ListByOrderID 查询指定订单的操作日志
func (r *RealTradeLogRepository) ListByOrderID(orderID string) ([]model.RealTradeLog, error) {
	var logs []model.RealTradeLog
	err := r.db.Where("order_id = ?", orderID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}
