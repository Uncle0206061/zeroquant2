// Package repository 提供实盘账户数据访问层
package repository

import (
	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
	"gorm.io/gorm"
)

// RealPortfolioRepository 实盘账户数据访问
type RealPortfolioRepository struct {
	db *gorm.DB
}

// NewRealPortfolioRepository 创建 RealPortfolioRepository
func NewRealPortfolioRepository(db *gorm.DB) *RealPortfolioRepository {
	return &RealPortfolioRepository{db: db}
}

// FindByUserID 根据用户ID查实盘账户
func (r *RealPortfolioRepository) FindByUserID(userID int64) (*model.RealPortfolio, error) {
	var portfolio model.RealPortfolio
	err := r.db.Where("user_id = ?", userID).First(&portfolio).Error
	return &portfolio, err
}

// Create 创建实盘账户
func (r *RealPortfolioRepository) Create(p *model.RealPortfolio) error {
	return r.db.Create(p).Error
}

// Save 保存实盘账户
func (r *RealPortfolioRepository) Save(p *model.RealPortfolio) error {
	return r.db.Save(p).Error
}
