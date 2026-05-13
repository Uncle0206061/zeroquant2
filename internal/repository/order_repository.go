// Package repository 提供数据访问层
package repository

import (
	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
	"gorm.io/gorm"
)

// OrderRepository 订单数据访问
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建 OrderRepository
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create 创建订单
func (r *OrderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

// Update 更新订单
func (r *OrderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

// FindByOrderID 根据委托单号查订单
func (r *OrderRepository) FindByOrderID(orderID string) (*model.Order, error) {
	var order model.Order
	err := r.db.Where("order_id = ?", orderID).First(&order).Error
	return &order, err
}

// FindByID 根据主键查订单
func (r *OrderRepository) FindByID(id int64) (*model.Order, error) {
	var order model.Order
	err := r.db.First(&order, id).Error
	return &order, err
}

// ListByUser 分页查询用户订单
func (r *OrderRepository) ListByUser(userID int64, status string, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}
