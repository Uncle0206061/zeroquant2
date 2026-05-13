// Package repository 提供数据访问层
package repository

import (
	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
	"gorm.io/gorm"
)

// PositionRepository 持仓数据访问
type PositionRepository struct {
	db *gorm.DB
}

// NewPositionRepository 创建 PositionRepository
func NewPositionRepository(db *gorm.DB) *PositionRepository {
	return &PositionRepository{db: db}
}

// FindByUserAndStock 查询用户指定股票的持仓
func (r *PositionRepository) FindByUserAndStock(userID int64, stockCode string) (*model.Position, error) {
	var pos model.Position
	err := r.db.Where("user_id = ? AND stock_code = ?", userID, stockCode).First(&pos).Error
	return &pos, err
}

// ListByUser 查询用户全部持仓
func (r *PositionRepository) ListByUser(userID int64) ([]model.Position, error) {
	var positions []model.Position
	err := r.db.Where("user_id = ? AND quantity > 0", userID).Order("created_at DESC").Find(&positions).Error
	return positions, err
}

// Save 保存持仓（新建或更新）
func (r *PositionRepository) Save(pos *model.Position) error {
	return r.db.Save(pos).Error
}

// Create 创建持仓
func (r *PositionRepository) Create(pos *model.Position) error {
	return r.db.Create(pos).Error
}
