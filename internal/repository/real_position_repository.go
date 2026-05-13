// Package repository 提供实盘持仓数据访问层
package repository

import (
	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
	"gorm.io/gorm"
)

// RealPositionRepository 实盘持仓数据访问
type RealPositionRepository struct {
	db *gorm.DB
}

// NewRealPositionRepository 创建 RealPositionRepository
func NewRealPositionRepository(db *gorm.DB) *RealPositionRepository {
	return &RealPositionRepository{db: db}
}

// FindByUserAndStock 查询用户指定股票的实盘持仓
func (r *RealPositionRepository) FindByUserAndStock(userID int64, stockCode string) (*model.RealPosition, error) {
	var pos model.RealPosition
	err := r.db.Where("user_id = ? AND stock_code = ?", userID, stockCode).First(&pos).Error
	return &pos, err
}

// ListByUser 查询用户全部实盘持仓
func (r *RealPositionRepository) ListByUser(userID int64) ([]model.RealPosition, error) {
	var positions []model.RealPosition
	err := r.db.Where("user_id = ? AND quantity > 0", userID).Order("created_at DESC").Find(&positions).Error
	return positions, err
}

// Save 保存实盘持仓（新建或更新）
func (r *RealPositionRepository) Save(pos *model.RealPosition) error {
	return r.db.Save(pos).Error
}

// Create 创建实盘持仓
func (r *RealPositionRepository) Create(pos *model.RealPosition) error {
	return r.db.Create(pos).Error
}
