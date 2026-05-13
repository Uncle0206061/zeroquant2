// Package repository 提供实盘订单数据访问层
package repository

import (
	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
	"gorm.io/gorm"
)

// RealOrderRepository 实盘订单数据访问
type RealOrderRepository struct {
	db *gorm.DB
}

// NewRealOrderRepository 创建 RealOrderRepository
func NewRealOrderRepository(db *gorm.DB) *RealOrderRepository {
	return &RealOrderRepository{db: db}
}

// Create 创建实盘订单
func (r *RealOrderRepository) Create(order *model.RealOrder) error {
	return r.db.Create(order).Error
}

// Update 更新实盘订单
func (r *RealOrderRepository) Update(order *model.RealOrder) error {
	return r.db.Save(order).Error
}

// FindByOrderID 根据委托单号查实盘订单
func (r *RealOrderRepository) FindByOrderID(orderID string) (*model.RealOrder, error) {
	var order model.RealOrder
	err := r.db.Where("order_id = ?", orderID).First(&order).Error
	return &order, err
}

// ListByUser 分页查询用户实盘订单
func (r *RealOrderRepository) ListByUser(userID int64, status string, page, pageSize int) ([]model.RealOrder, int64, error) {
	var orders []model.RealOrder
	var total int64

	query := r.db.Model(&model.RealOrder{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}
