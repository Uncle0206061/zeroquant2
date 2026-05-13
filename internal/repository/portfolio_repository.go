// Package repository 提供数据访问层
package repository

import (
	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
	"gorm.io/gorm"
)

// PortfolioRepository 模拟账户数据访问
type PortfolioRepository struct {
	db *gorm.DB
}

// NewPortfolioRepository 创建 PortfolioRepository
func NewPortfolioRepository(db *gorm.DB) *PortfolioRepository {
	return &PortfolioRepository{db: db}
}

// FindByUserID 根据用户ID查模拟账户
func (r *PortfolioRepository) FindByUserID(userID int64) (*model.Portfolio, error) {
	var portfolio model.Portfolio
	err := r.db.Where("user_id = ?", userID).First(&portfolio).Error
	return &portfolio, err
}

// Create 创建模拟账户
func (r *PortfolioRepository) Create(p *model.Portfolio) error {
	return r.db.Create(p).Error
}

// Save 保存账户
func (r *PortfolioRepository) Save(p *model.Portfolio) error {
	return r.db.Save(p).Error
}
